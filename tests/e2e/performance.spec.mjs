// SPDX-License-Identifier: EUPL-1.2
import { test,expect } from './helpers/studio.mjs';
import { referenceSource,assess } from './helpers/performance.mjs';
import { mkdir,writeFile,readFile,readdir } from 'node:fs/promises';
import { cpus,totalmem,platform,release,arch,availableParallelism } from 'node:os';
import { createHash } from 'node:crypto';
import { join } from 'node:path';
import { execFile } from 'node:child_process';
import { promisify } from 'node:util';
import { root } from './helpers/fixtures.mjs';

test.use({studioWorkload:true});
const visibleCurrent=async(page,text)=>{
  await page.waitForFunction(expected=>{
    const card=document.querySelector('#ref-live-card'),status=document.querySelector('#studio-status');
    const paragraph=card?.querySelector('p'),rect=paragraph?.getBoundingClientRect();
    return rect?.width>0 && rect.height>0 && rect.top<innerHeight && rect.bottom>0 &&
      getComputedStyle(paragraph).visibility==='visible' && paragraph.textContent.includes(expected) &&
      status?.querySelector('strong')?.textContent==='current' && status.querySelector('div')?.textContent==='' &&
      !document.querySelector('#app-search')?.disabled;
  },text,{polling:'raf',timeout:15_000});
  await page.evaluate(()=>new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve))));
};

test('T049 reference-machine automated acceptance',async({page,context,studio,tools,browser},testInfo)=>{
  const updates=[],startups=[],sequences=[],sequenceEvidence=[];
  const report={schema:1,startedUTC:new Date().toISOString(),machine:{os:platform(),release:release(),arch:arch(),cpu:cpus()[0]?.model,logicalCPUs:availableParallelism(),memoryBytes:totalmem()},
    tools:{node:process.versions.node,chromium:browser.version(),playwright:'1.64.0',go:tools.records.find(r=>r.args[0]==='version'&&r.executable==='go')?.stdout.trim(),hugo:tools.records.find(r=>r.executable===tools.hugo&&r.args[0]==='version')?.stdout.trim()},
    method:{viewport:'1280x800',startup:'process spawn to rendered current card and enabled search, including bootstrap, navigation and initial build',edit:'external writeFile completion to rendered current card and enabled search',clock:'Node performance.now, viewport-intersecting browser RAF predicate plus two frames (paint opportunity, not a pixel timestamp)',warmup:'none discarded; fresh process and ephemeral port each start; browser and development CLI built before timing; OS caches not flushed',retries:0},
    workload:{sheets:100,bytesPerSheet:8192,maximumAllowedBytesPerSheet:102400},sequenceEvidence};
  const evidence=testInfo.outputPath('performance.json');
  try {
    for(let i=0;i<20;i++){
      if(i>0)await page.goto('about:blank');
      const startedAt=i===0?studio.startedAt:(await studio.restart({newPort:true})).startedAt;
      await page.goto(studio.origin+'/sample/#ref-live-card');await visibleCurrent(page,'Initial content');
      startups.push(performance.now()-startedAt);
      if(i===19)console.log('T049: all 20 startup samples recorded');
      expect((await studio.readState()).sheets.filter(s=>s.status==='current')).toHaveLength(100);
    }
    const git=promisify(execFile);
    report.source={revision:(await git('git',['rev-parse','HEAD'],{cwd:root,timeout:5000})).stdout.trim()};
    await git('git',['diff','--exit-code','HEAD','--','cmd','internal','assets','go.mod','go.sum','tests/e2e'],{cwd:root,timeout:5000});
    report.source.runtimeAndHarnessClean=true;
    const names=(await readdir(studio.directory)).filter(n=>n.endsWith('.md')).sort();expect(names).toHaveLength(100);
    const digest=createHash('sha256');
    for(const name of names){const bytes=await readFile(join(studio.directory,name));expect(bytes.length).toBe(8192);digest.update(name+'\0').update(bytes).update('\0')}
    report.workload.initialSHA256=digest.digest('hex');
    for(let i=0;i<100;i++){
      const text='Measured saved edit '+i;
      await writeFile(studio.path,referenceSource(0,text));const savedAt=performance.now();
      await visibleCurrent(page,text);updates.push(performance.now()-savedAt);
      if((i+1)%25===0)console.log(`T049: ${i+1}/100 edit samples recorded`);
    }
    let lastValid='Measured saved edit 99';
    for(let i=0;i<100;i++){
      const invalid=i%2===1,disconnected=i===50;
      if(disconnected){await context.setOffline(true);await expect(page.locator('#studio-status > strong')).toHaveText('Disconnected — freshness unknown')}
      await writeFile(studio.path,referenceSource(0,invalid?'[Obsolete](javascript:alert(1))':'Obsolete revision '+i));
      await writeFile(studio.path,'invalid intermediate');
      const text='Latest sequence '+i;
      const final=referenceSource(0,invalid?'\n'.repeat(i+1)+'[Invalid latest](javascript:alert(1))':text);
      await writeFile(studio.path,final);
      const expectedLine=invalid?final.split('\n').findIndex(line=>line.includes('[Invalid latest]'))+1:null;
      if(disconnected){
        await expect.poll(async()=> (await studio.readState()).sheets.find(s=>s.name==='sample.md')?.status).toBe('current');
        await expect(page.locator('#studio-status > strong')).toHaveText('Disconnected — freshness unknown');
        await context.setOffline(false);
      }
      if(invalid){
        await expect(page.locator('#studio-status > strong')).toHaveText('invalid — last-valid preview');
        await expect(page.locator('#studio-status')).toContainText(`sample.md:${expectedLine}:`);
        await expect(page.locator('#ref-live-card')).toContainText(lastValid);
      }else{await visibleCurrent(page,text);lastValid=text}
      const state=(await studio.readState()).sheets.find(s=>s.name==='sample.md');
      const correct=invalid?state.status==='invalid'&&state.stale&&state.diagnostics.length===1&&state.diagnostics[0].code==='REN002'&&state.diagnostics[0].span.start.line===expectedLine:state.status==='current'&&!state.stale&&(state.diagnostics?.length??0)===0;
      sequenceEvidence.push({index:i,finalStatus:invalid?'invalid':'current',expectedLine,disconnected,correct});sequences.push(correct);
      expect(await readFile(studio.path,'utf8')).toBe(final);expect(correct).toBe(true);
      if((i+1)%25===0)console.log(`T049: ${i+1}/100 latest-save sequences checked`);
    }
  } finally {
    report.completedUTC=new Date().toISOString();report.assessment=assess({updates,startups,sequences});
    await mkdir(testInfo.outputDir,{recursive:true});await writeFile(evidence,JSON.stringify(report,null,2)+'\n');
    // No cookies, bootstrap URLs, authenticated traces, source paths or hostnames.
    console.log('T049 '+JSON.stringify({edits:report.assessment.edits.withinThreshold,startups:report.assessment.startups.withinThreshold,sequences:report.assessment.latestSave.correct,passed:report.assessment.passed,evidence}));
  }
  expect(report.assessment.passed,'T049 thresholds; every raw sample retained in local performance.json').toBe(true);
});
