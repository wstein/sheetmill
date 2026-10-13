// SPDX-License-Identifier: EUPL-1.2
// Controlled fault-injection process, never a source-sheet command.
import { appendFile } from 'node:fs/promises';
import { OwnedWorkspace, registerInterruptCleanup } from './process.mjs';
const owner=await OwnedWorkspace.create();
registerInterruptCleanup(owner);
const evidence=process.argv[2];
owner.defer(()=>appendFile(evidence,'cleaned\n'));
let extra;
if(process.argv.includes('--two')) {
  extra=await OwnedWorkspace.create();registerInterruptCleanup(extra);
  extra.defer(()=>new Promise(resolve=>setTimeout(resolve,250)));
}
const pending=owner.run(process.execPath,['-e','setInterval(()=>{},1000)']);
pending.catch(()=>{});
await appendFile(evidence,'prepared\n');
process.stdout.write((extra?JSON.stringify([owner.root,extra.root]):owner.root)+'\n');
