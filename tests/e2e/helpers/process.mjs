// SPDX-License-Identifier: EUPL-1.2
import { spawn } from 'node:child_process';
import { mkdtemp, realpath, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';

export function validateVersions(v) {
  const checks = {
    node: v.node === '24.21.0', playwright: v.playwright === '1.64.0',
    chromium: v.chromium === '156.0.8078.4',
    go: /^go version go1\.27\.2\s/.test(v.go || ''),
    hugo: /^hugo v0\.167\.0(?:[+\s-]|$)/.test(v.hugo || ''),
  };
  for (const [tool, valid] of Object.entries(checks)) if (!valid) throw Error(`E2E requires pinned ${tool}; got ${v[tool] ?? 'missing'}`);
}
const redact = text => text.replace(/(authorization|cookie|token|secret)([=: ]+)[^\s]+/gi, '$1$2[redacted]');

// Every spawned process owns a fresh process group. Deadline/abort escalation
// targets only this group, including descendants created by the Go compiler.
export function run(executable, args, { cwd, env = process.env, timeoutMs = 15_000, maxBytes = 1024*1024, signal, onRecord = () => {} } = {}) {
  if (signal?.aborted) return Promise.reject(Error('operation aborted'));
  return new Promise((resolve, reject) => {
    const child = spawn(executable, args, { cwd, env, shell: false, detached: true, stdio: ['ignore','pipe','pipe'] });
    let stdout=Buffer.alloc(0), stderr=Buffer.alloc(0), reason, spawnError, killTimer;
    const kill = sig => {
      if (!child.pid) return;
      try { process.kill(-child.pid,sig); } catch (err) { if (err.code !== 'ESRCH') spawnError ??= err; }
    };
    const stop = why => {
      if (reason) return;
      reason=why; kill('SIGTERM');
      killTimer=setTimeout(() => kill('SIGKILL'),1000);
    };
    const capture = which => chunk => {
      const remaining=Math.max(0,maxBytes-stdout.length-stderr.length);
      if (which==='out') stdout=Buffer.concat([stdout,chunk.subarray(0,remaining)]);
      else stderr=Buffer.concat([stderr,chunk.subarray(0,remaining)]);
      if (chunk.length>remaining) stop('output limit exceeded');
    };
    child.stdout.on('data',capture('out')); child.stderr.on('data',capture('err'));
    child.on('error',err=>{spawnError=err;});
    const abort=()=>stop('operation aborted');
    signal?.addEventListener('abort',abort,{once:true});
    const timer=setTimeout(()=>stop('command deadline exceeded'),timeoutMs);
    child.on('close',(code,termination)=>{
      clearTimeout(timer);clearTimeout(killTimer);signal?.removeEventListener('abort',abort);
      const record={ executable, args:[...args], code, termination, stdout:redact(stdout.toString()), stderr:redact(stderr.toString()), reason:reason||spawnError?.message };
      try { onRecord(record); } catch (err) { reject(err);return; }
      if (reason||spawnError||code!==0) reject(Error(`${reason||spawnError?.message||`exit ${code}`}: ${record.stderr}`));
      else resolve(record);
    });
  });
}

export class OwnedWorkspace {
  static async create() { return new OwnedWorkspace(await mkdtemp(join(await realpath(tmpdir()),'sheetmill-e2e-'))); }
  constructor(root) { this.root=root;this.controller=new AbortController();this.pending=new Set();this.disposers=[]; }
  run(executable,args,options={}) {
    const result=run(executable,args,{...options,signal:this.controller.signal});
    this.pending.add(result);result.then(()=>this.pending.delete(result),()=>this.pending.delete(result));
    return result;
  }
  defer(dispose) { this.disposers.push(dispose); }
  close() {
    if (this.closing) return this.closing;
    this.closing=(async()=>{
      this.controller.abort();
      await Promise.allSettled([...this.pending]);
      const errors=[];
      for (const dispose of this.disposers.reverse()) try { await dispose(); } catch (e) { errors.push(e); }
      await rm(this.root,{recursive:true,force:true});
      if (errors.length) throw new AggregateError(errors,'owned resource cleanup failed');
    })();
    return this.closing;
  }
}

export function registerInterruptCleanup(owner) {
  const handlers=new Map();
  for (const signal of ['SIGINT','SIGTERM']) {
    const handler=()=>{owner.close().finally(()=>process.exit(signal==='SIGINT'?130:143));};
    handlers.set(signal,handler);process.once(signal,handler);
  }
  const remove=()=>{for(const [s,h] of handlers)process.removeListener(s,h);};
  owner.defer(remove);
  return remove;
}
