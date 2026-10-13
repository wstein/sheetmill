// SPDX-License-Identifier: EUPL-1.2
// Fixed, independently authored reference input; commands are display text only.
export function referenceSource(index, text) {
  const prefix=`---\ntitle: Performance reference ${String(index).padStart(3,'0')}\n---\n## Workflow\n### Live card {#live-card}\n${text}\n\n### Commands\n\n\`\`\`text\nsmill check reference.md\n\`\`\`\n\n### Notes\n\n`;
  const padding='Reference notes for the local preview workload. ';
  const remaining=8192-Buffer.byteLength(prefix)-1;
  if(remaining<0)throw Error('Reference input exceeds fixed workload size');
  return prefix+padding.repeat(Math.ceil(remaining/padding.length)).slice(0,remaining)+'\n';
}
export function summarize(values, thresholdMs) {
  const valid=Array.isArray(values)&&values.length>0&&values.every(v=>Number.isFinite(v)&&v>=0);
  const sorted=valid?[...values].sort((a,b)=>a-b):[];
  const percentile=p=>valid?sorted[Math.ceil(sorted.length*p)-1]:null;
  return {valid,count:values.length,thresholdMs,withinThreshold:valid?values.filter(v=>v<=thresholdMs).length:0,
    p50Ms:percentile(.5),p95Ms:percentile(.95),maxMs:valid?sorted.at(-1):null,samplesMs:[...values]};
}
export function assess({updates,startups,sequences}) {
  const edit=summarize(updates,1000),start=summarize(startups,3000);
  const correct=sequences.filter(v=>v===true).length;
  return {passed:edit.valid&&edit.count===100&&edit.withinThreshold>=95&&start.valid&&start.count===20&&start.withinThreshold>=19&&sequences.length===100&&correct===100,
    edits:edit,startups:start,latestSave:{count:sequences.length,correct,results:[...sequences]},usability:{status:'pending',requiredAuthors:5}};
}
