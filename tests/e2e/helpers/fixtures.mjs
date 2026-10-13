// SPDX-License-Identifier: EUPL-1.2
import { test as base, expect } from '@playwright/test';
import { mkdir, readFile, writeFile, access } from 'node:fs/promises';
import { appendFileSync } from 'node:fs';
import { join, isAbsolute } from 'node:path';
import { fileURLToPath } from 'node:url';
import { createHash } from 'node:crypto';
import { OwnedWorkspace, registerInterruptCleanup, validateVersions, toolEnvironment } from './process.mjs';
import { loadCorpus } from './corpus.mjs';
import { readSite, compareSites, serveSite } from './site.mjs';
import { guardNetwork } from './network.mjs';

export const root=fileURLToPath(new URL('../../../',import.meta.url));
const hash=bytes=>createHash('sha256').update(bytes).digest('hex');
async function evidence() {
  const directory=join(root,'.local/e2e/runs');await mkdir(directory,{recursive:true});
  const {mkdtemp}=await import('node:fs/promises');return mkdtemp(join(directory,'run-'));
}
function recordTo(path,records) {return record=>{records.push(record);appendFileSync(path,JSON.stringify(record)+'\n');};}

export const test=base.extend({
  sourceSuffix:['',{option:true}],
  basePath:['/',{option:true}],
  corpusID:['smill',{option:true}],
  corpusCases:[async({},use)=>{
    await use(await loadCorpus(root,JSON.parse(await readFile(join(root,'tests/e2e/corpus.json'),'utf8')),process.env.SHEETMILL_E2E_CORPUS??'smill'));
  },{scope:'worker'}],
  corpus:async({corpusCases,corpusID},use)=>{
    const item=corpusCases.find(c=>c.id===corpusID);
    if(!item)throw Error('Requested corpus case is not selected: '+corpusID);
    await use(item);
  },
  tools:[async({browser},use)=>{
    const directory=await evidence(),log=join(directory,'tools.jsonl'),records=[];
    await writeFile(log,'');
    const owner=await OwnedWorkspace.create();registerInterruptCleanup(owner);
    const env=toolEnvironment(join(owner.root,'go-cache'));
    const exec=(executable,args,options={})=>owner.run(executable,args,{env,...options,onRecord:recordTo(log,records)});
    try {
      const hugo=process.env.SHEETMILL_HUGO;
      if(!hugo||!isAbsolute(hugo))throw Error('SHEETMILL_HUGO must name an absolute installed Hugo 0.167.0 executable');
      const go=(await exec('go',['version'])).stdout;
      const hv=(await exec(hugo,['version'])).stdout;
      const pkg=JSON.parse(await readFile(new URL('../node_modules/@playwright/test/package.json',import.meta.url),'utf8'));
      validateVersions({node:process.versions.node,playwright:pkg.version,chromium:browser.version(),go,hugo:hv});
      const smill=join(owner.root,'smill');
      await exec('go',['build','-trimpath','-o',smill,'./cmd/smill'],{cwd:root,env,timeoutMs:120_000});
      const version=await exec(smill,['version'],{env});expect(version.stdout).toBe('smill dev\n');
      await use({smill,hugo,env,log,records});
    } catch(error) {
      appendFileSync(log,JSON.stringify({stage:'tool-setup',error:error.message})+'\n');
      throw Error(`${error.message} (setup evidence: ${log})`,{cause:error});
    } finally {await owner.close();}
  },{scope:'worker',timeout:300_000}],
  site:async({tools,corpus,corpusCases,sourceSuffix,basePath},use,testInfo)=>{
    const directory=await evidence(),log=join(directory,'stages.jsonl'),records=[];
    await writeFile(log,'');
    const owner=await OwnedWorkspace.create();registerInterruptCleanup(owner);
    const sources=[];
    const exec=(executable,args,options={})=>owner.run(executable,args,{env:tools.env,cwd:owner.root,...options,onRecord:recordTo(log,records)});
    try {
      const input=join(owner.root,'sheets');await mkdir(input);
      for(const item of corpusCases) {
        const source=join(input,item.id+'.md');
        await writeFile(source,Buffer.concat([item.sourceBytes,Buffer.from(item.id===corpus.id?sourceSuffix:'')]));
        sources.push({item,source,sourceHash:hash(await readFile(source)),originalHash:hash(item.sourceBytes)});
        const checked=await exec(tools.smill,['check',source]);expect(checked.stdout).toBe('');
        const inspected=JSON.parse((await exec(tools.smill,['inspect',source])).stdout);
        expect(inspected.schema).toBe(1);expect(inspected.metadata.title).toBe(item.title);
        expect(inspected.metadata.source.revision).toBe(item.product.revision);
        expect(inspected.metadata.source.license).toBe(item.attribution.license);
        expect(inspected.metadata.source.attribution).toBe(item.attribution.author);
        const ids=inspected.sections.flatMap(s=>s.cards.map(c=>c.id));
        for(const id of item.cards)expect(ids).toContain(id);
      }
      const source=corpusCases.length===1?sources[0].source:input;
      const sites=[];
      for(const name of ['first','second']) {
        const project=join(owner.root,name),destination=join(project,'public');
        await exec(tools.smill,['build','hugo',source,'--out',project,'--base-path',basePath]);
        await exec(tools.hugo,['--source',project,'--destination',destination,'--clock','2000-01-01T00:00:00Z','--noBuildLock','--quiet'],{timeoutMs:30_000,env:{...tools.env,HUGO_CACHEDIR:join(owner.root,'cache-'+name)}});
        sites.push(await readSite(destination));
      }
      compareSites(...sites);
      const digest=createHash('sha256');for(const [name,bytes] of sites[0])digest.update(name+'\0').update(bytes).update('\0');
      await writeFile(join(directory,'result.json'),JSON.stringify({files:sites[0].size,sha256:digest.digest('hex'),sourceHashes:Object.fromEntries(sources.map(s=>[s.item.id,s.sourceHash]))},null,2));
      const mounted=basePath==='/'?sites[0]:new Map([...sites[0]].map(([name,bytes])=>[basePath.slice(1)+name,bytes]));
      const server=await serveSite(mounted);owner.defer(server.close);
      await use({...server,records,toolRecords:tools.records,workspace:owner.root,log});
    } finally {
      try {
        for(const {item,source,sourceHash,originalHash} of sources) {
          expect(hash(await readFile(source))).toBe(sourceHash);
          expect(hash(await readFile(item.sourcePath))).toBe(originalHash);
        }
        await expect(access(join(owner.root,'SHEETMILL_SNIPPET_MUST_NOT_RUN'))).rejects.toThrow();
      } finally {
        await owner.close();
        await testInfo.attach('tool-stages',{path:tools.log,contentType:'application/jsonl'});
        await testInfo.attach('site-stages',{path:log,contentType:'application/jsonl'});
      }
    }
  },
  _network:[async({context,site},use)=>{
    const violations=await guardNetwork(context,site.origin);
    await use();expect(violations,'off-origin browser requests').toEqual([]);
  },{auto:true}],
});
export {expect};
