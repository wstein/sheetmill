// SPDX-License-Identifier: EUPL-1.2
import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, realpath, writeFile, symlink, access, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { run, OwnedWorkspace, validateVersions } from './helpers/process.mjs';
import { readSite, compareSites, serveSite } from './helpers/site.mjs';

const scratch = async () => mkdtemp(join(await realpath(tmpdir()), 'sheetmill-helper-'));
test('arguments are literal; no shell interpolation', async () => {
  const root = await scratch();
  try {
    const literal = '$(touch ' + join(root, 'never') + ')';
    const result = await run(process.execPath, ['-e', 'process.stdout.write(process.argv[1])', literal]);
    assert.equal(result.stdout, literal);
    await assert.rejects(access(join(root, 'never')));
  } finally { await rm(root, { recursive: true }); }
});
test('missing executables and failed commands are explicit', async () => {
  await assert.rejects(run('/missing/sheetmill-test-tool', []), /ENOENT/);
  await assert.rejects(run(process.execPath, ['-e', 'process.stderr.write("failed");process.exit(7)']), /exit 7.*failed/s);
});
test('deadline kills even a SIGTERM-resistant owned child', async () => {
  const start = Date.now();
  await assert.rejects(run(process.execPath, ['-e', 'process.on("SIGTERM",()=>{});setInterval(()=>{},1000)'], { timeoutMs: 300 }), /deadline/);
  assert.ok(Date.now() - start < 4000);
});
test('output bound is enforced across stdout and stderr', async () => {
  let record;
  await assert.rejects(run(process.execPath, ['-e', 'process.stdout.write("x".repeat(2000));setInterval(()=>{},1000)'], { maxBytes: 100, onRecord: r => { record = r; } }), /output limit/);
  assert.ok(Buffer.byteLength(record.stdout + record.stderr) <= 100);
});
test('aborted operation does not start work', async () => {
  const controller = new AbortController(); controller.abort();
  await assert.rejects(run(process.execPath, ['-e', 'process.exit(0)'], { signal: controller.signal }), /aborted/);
});
test('exact tool versions, standard and extended Hugo forms', () => {
  const versions = { node: '24.21.0', playwright: '1.64.0', chromium: '156.0.8078.4', go: 'go version go1.27.2 darwin/arm64', hugo: 'hugo v0.167.0-abcdef linux/amd64' };
  validateVersions(versions);
  validateVersions({ ...versions, hugo: 'hugo v0.167.0+extended darwin/arm64' });
  for (const [key,value] of Object.entries({ node:'26.11.1', playwright:'1.63.0', chromium:'155.0.0.0', go:'go version go1.27.20 linux/amd64', hugo:'hugo v0.167.01 linux/amd64' })) {
    assert.throws(() => validateVersions({ ...versions, [key]:value }), /requires/);
  }
});
test('all site names and bytes matter; symlinks are refused', async () => {
  const root = await scratch();
  try {
    await writeFile(join(root, 'index.html'), 'first');
    const first = await readSite(root);
    compareSites(first, new Map(first));
    assert.throws(() => compareSites(first, new Map()), /file list/);
    assert.throws(() => compareSites(first, new Map([['index.html', Buffer.from('other')]])), /bytes/);
    await symlink(join(root,'index.html'),join(root,'link.html'));
    await assert.rejects(readSite(root), /regular/);
  } finally { await rm(root, { recursive: true }); }
});
test('owned server maps only generated files and closes its listener', async () => {
  const server = await serveSite(new Map([['index.html',Buffer.from('home')],['smill/index.html',Buffer.from('sheet')]]));
  assert.match(server.origin, /^http:\/\/127\.0\.0\.1:\d+$/);
  assert.equal(await (await fetch(server.origin+'/smill/')).text(),'sheet');
  for (const path of ['/unknown','/%ZZ','/%2e%2e/secret','/index.html?x=1']) {
    const response=await fetch(server.origin+path);
    assert.equal(response.status,path.includes('?')?200:404);
  }
  assert.equal((await fetch(server.origin,{method:'POST'})).status,405);
  await server.close();
  await assert.rejects(fetch(server.origin));
});
test('failed setup cleanup removes only owned workspace', async () => {
  const owner = await OwnedWorkspace.create(); const root=owner.root;
  await assert.rejects(owner.run('/missing/sheetmill-test-tool', []));
  await owner.close(); await owner.close();
  await assert.rejects(access(root));
});
test('workspace cancellation awaits child shutdown before removal', async () => {
  const owner=await OwnedWorkspace.create();
  const pending=owner.run(process.execPath,['-e','setInterval(()=>{},1000)']);
  const failed=assert.rejects(pending,/aborted/);
  await owner.close(); await failed;
  await assert.rejects(access(owner.root));
});
