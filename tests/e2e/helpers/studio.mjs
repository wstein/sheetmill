// SPDX-License-Identifier: EUPL-1.2
import { get } from 'node:http';
import { spawn } from 'node:child_process';
import { once } from 'node:events';
import { mkdir,writeFile,readFile } from 'node:fs/promises';
import { join } from 'node:path';
import { test as publishing,expect } from './fixtures.mjs';
import { OwnedWorkspace,registerInterruptCleanup } from './process.mjs';
import { guardNetwork } from './network.mjs';
import { referenceSource } from './performance.mjs';

export const source=text=>`---\ntitle: Live reference\n---\n## Workflow\n### Live card {#live-card}\n${text}\n`;
async function server(owner,tools,input,port=0){
  const startedAt=performance.now();
  const child=spawn(tools.smill,['serve',input,'--hugo',tools.hugo,'--port',String(port)],{cwd:owner.root,env:tools.env,shell:false,detached:true,stdio:['ignore','pipe','pipe']});
  let closed=false;const exit=once(child,'close').then(([code])=>{closed=true;return code});
  let stopping;
  const stop=()=>stopping??=(async()=>{
    if(!closed){try{process.kill(-child.pid,'SIGTERM')}catch(e){if(e.code!=='ESRCH')throw e}}
    const timeout=setTimeout(()=>{if(!closed)try{process.kill(-child.pid,'SIGKILL')}catch{}},5000);
    try{const code=await exit;if(code!==0)throw Error('Serve did not exit gracefully')}finally{clearTimeout(timeout)}
  })();owner.defer(stop);
  let output='',bytes=0;
  const ready=new Promise((resolve,reject)=>{
    child.on('error',()=>reject(Error('Serve executable could not start')));
    child.stderr.on('data',chunk=>{bytes+=chunk.length;if(bytes>1024*1024){stop().catch(()=>{});reject(Error('Serve output limit exceeded'))}});
    child.stdout.on('data',chunk=>{bytes+=chunk.length;if(bytes>1024*1024){stop().catch(()=>{});reject(Error('Serve output limit exceeded'));return};output+=chunk.toString();const match=output.match(/Open: (http:\/\/127\.0\.0\.1:\d+\/studio\/bootstrap\?token=[a-f0-9]{64})/);if(match)resolve(match[1])});
    exit.then(()=>reject(Error('Serve exited before readiness')));
  });
  let timer;try{const bootstrap=await Promise.race([ready,new Promise((_,reject)=>{timer=setTimeout(()=>reject(Error('Serve readiness deadline exceeded')),120000)})]);return {origin:new URL(bootstrap).origin,bootstrap,stop,exit,startedAt}}finally{clearTimeout(timer)}
}
export const test=publishing.extend({
  studioSingle:[false,{option:true}],studioInvalid:[false,{option:true}],studioWorkload:[false,{option:true}],
  studio:async({tools,context,studioSingle,studioInvalid,studioWorkload},use)=>{
    const owner=await OwnedWorkspace.create();registerInterruptCleanup(owner);
    const directory=join(owner.root,'sheets');
    try{
    await mkdir(directory);
    const path=join(directory,'sample.md'),peer=join(directory,'private.md');
    await writeFile(path,studioWorkload?referenceSource(0,'Initial content'):studioInvalid?'invalid':source('Initial content'));
    if(studioWorkload) for(let i=1;i<100;i++)await writeFile(join(directory,'sheet'+String(i).padStart(3,'0')+'.md'),referenceSource(i,'Initial peer '+i));
    if(studioSingle)await writeFile(peer,source('UNSELECTED_PRIVATE_CONTENT'));
      let running=await server(owner,tools,studioSingle?path:directory);
      // Consume bootstrap outside browser history/reporting. Never retain raw
      // startup stdout, token URLs, cookies or authenticated Playwright traces.
      let cookie,cookieName;
      const authenticate=async()=>{try{
        const response=await new Promise((resolve,reject)=>{get(running.bootstrap,{headers:{'Sec-Fetch-Site':'none','Sec-Fetch-Mode':'navigate','Sec-Fetch-Dest':'document'}},res=>{res.resume();resolve(res)}).on('error',reject)});
        if(response.statusCode!==303||response.headers.location!=='/')throw Error('Bootstrap response refused');
        const pair=response.headers['set-cookie'][0].split(';')[0];cookieName=pair.split('=')[0];cookie=pair.split('=').slice(1).join('=');
        await context.addCookies([{name:cookieName,value:cookie,url:running.origin,httpOnly:true,sameSite:'Strict'}]);
      }catch{throw Error('Local preview bootstrap failed (credentials omitted)')}};
      await authenticate();
      const readState=async()=>{
        const response=await fetch(running.origin+'/studio/state',{headers:{Cookie:cookieName+'='+cookie}});expect(response.status).toBe(200);return response.json();
      };
      const restart=async({newPort=false}={})=>{const port=newPort?0:Number(new URL(running.origin).port);await running.stop();running=await server(owner,tools,studioSingle?path:directory,port);await authenticate();return {startedAt:running.startedAt}};
      await use({...running,get origin(){return running.origin},bootstrap:undefined,path,peer,directory,readState,restart});
    }finally{await owner.close()}
  },
  _network:[async({context,studio},use)=>{
    const violations=await guardNetwork(context,()=>studio.origin);await use();expect(violations).toEqual([]);
  },{auto:true}],
});
export {expect};
