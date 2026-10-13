// SPDX-License-Identifier: EUPL-1.2
import test from 'node:test';
import assert from 'node:assert/strict';
import { referenceSource, summarize, assess } from './helpers/performance.mjs';
import { guardNetwork } from './helpers/network.mjs';

test('reference workload has 100 distinct titles and exactly 8 KiB per sheet',()=>{
  const titles=new Set();
  for(let i=0;i<100;i++){
    const text=referenceSource(i,'Latest edit '+i);
    assert.equal(Buffer.byteLength(text),8192);
    titles.add(text.split('\n')[1]);
    assert.ok(text.includes('### Live card {#live-card}'));
  }
  assert.equal(titles.size,100);
});
test('nearest-rank summaries preserve slow samples and exact threshold boundaries',()=>{
  const values=[...Array(95).fill(1000),...Array(5).fill(1001)];
  const result=summarize(values,1000);
  assert.equal(result.withinThreshold,95);assert.equal(result.p95Ms,1000);assert.equal(result.maxMs,1001);
  assert.deepEqual(result.samplesMs,values);
  assert.equal(assess({updates:values,startups:[...Array(19).fill(3000),3001],sequences:Array(100).fill(true)}).passed,true);
  assert.equal(assess({updates:[...Array(94).fill(1000),...Array(6).fill(1001)],startups:Array(20).fill(10),sequences:Array(100).fill(true)}).passed,false);
});
test('partial, malformed and incorrect sequence evidence cannot pass',()=>{
  const valid={updates:Array(100).fill(1),startups:Array(20).fill(1),sequences:Array(100).fill(true)};
  for(const update of [{updates:[]},{startups:Array(19).fill(1)},{sequences:Array(99).fill(true)},{updates:[NaN,...Array(99).fill(1)]},{updates:[-1,...Array(99).fill(1)]},{sequences:[false,...Array(99).fill(true)]}]) {
    assert.equal(assess({...valid,...update}).passed,false);
  }
});

test('changing session ports keeps the exact-origin network boundary',async()=>{
 let handler,origin='http://127.0.0.1:1111';
 const violations=await guardNetwork({route:async(pattern,callback)=>{handler=callback}},()=>origin);
 const results=[];
 const visit=url=>handler({request:()=>({url:()=>url}),continue:async()=>results.push('allowed'),abort:async()=>results.push('blocked')});
 await visit(origin+'/studio/state');origin='http://127.0.0.1:2222';
 await visit(origin+'/assets/app.js');await visit('http://127.0.0.1:1111/studio/state');await visit('https://example.invalid/');
 assert.deepEqual(results,['allowed','allowed','blocked','blocked']);assert.equal(violations.length,2);
});
