// SPDX-License-Identifier: EUPL-1.2
import { test,expect,source } from './helpers/studio.mjs';
import { writeFile,rename,unlink,symlink,readFile,access } from 'node:fs/promises';
import { join } from 'node:path';

const live=page=>page.locator('#ref-live-card');
const current=async(page,text)=>{await expect(live(page)).toContainText(text);await expect(page.locator('#studio-status > strong')).toHaveText('current')};
test('external saves, invalid recovery, peers, atomic replacement and removal',async({page,studio})=>{
  await page.goto(studio.origin+'/sample/#ref-live-card');await current(page,'Initial content');
  await writeFile(studio.path,source('Saved content'));await current(page,'Saved content');await expect(page).toHaveURL(studio.origin+'/sample/#ref-live-card');
  await writeFile(studio.path,'invalid');await expect(page.locator('#studio-status > strong')).toHaveText('invalid — last-valid preview');await expect(live(page)).toContainText('Saved content');await expect(page.locator('#studio-status')).toContainText('sample.md:1:1:');
  await writeFile(studio.path,source('[unsafe](javascript:alert(1))'));await expect(page.locator('#studio-status')).toContainText('unsafe');
  await writeFile(studio.peer,source('Valid peer'));await expect.poll(async()=> (await studio.readState()).sheets.find(s=>s.name==='private.md')?.status).toBe('current');
  await page.goto(studio.origin+'/private/');await current(page,'Valid peer');
  await writeFile(studio.path,source('Repaired content'));await page.goto(studio.origin+'/sample/#ref-live-card');await current(page,'Repaired content');
  const replacement=studio.path+'.new';await writeFile(replacement,source('Atomic replacement'));await rename(replacement,studio.path);await current(page,'Atomic replacement');
  await unlink(studio.peer);await expect.poll(async()=> (await studio.readState()).sheets.find(s=>s.name==='private.md')?.status).toBe('removed');
  await page.goto(studio.origin+'/private/');await expect(page.getByRole('heading',{name:'Reference unavailable'})).toBeVisible();
  await page.goto(studio.origin+'/sample/');await current(page,'Atomic replacement');
  await studio.stop();expect(await studio.exit).toBe(0);
});

test('100 rapid-save sequences converge and reconnect fetches full current state',async({page,context,studio})=>{
  await page.goto(studio.origin+'/sample/#ref-live-card');await current(page,'Initial content');
  for(let i=0;i<100;i++){
    await writeFile(studio.path,source('Obsolete '+i));await writeFile(studio.path,source('Newest sequence '+i));
    await current(page,'Newest sequence '+i);
  }
  await studio.restart();await page.reload();await current(page,'Newest sequence 99');
  await writeFile(studio.path,source('Same-port restart updated'));await current(page,'Same-port restart updated');
  await context.setOffline(true);await expect(page.locator('#studio-status > strong')).toHaveText('Disconnected — freshness unknown');
  await writeFile(studio.path,source('Saved while disconnected'));await expect.poll(async()=> (await studio.readState()).sheets[0].status).toBe('current');
  await context.setOffline(false);await current(page,'Saved while disconnected');await expect(page).toHaveURL(studio.origin+'/sample/#ref-live-card');
});

test.describe('single-file privacy and inert content',()=>{
  test.use({studioSingle:true});
  test('rejects private reads, source escapes, forged requests and executable content',async({page,studio})=>{
    expect((await fetch(studio.origin+'/studio/state')).status).toBe(403);
    await page.goto(studio.origin+'/sample/');await current(page,'Initial content');
    const state=await studio.readState();expect(state.sheets.map(s=>s.name)).toEqual(['sample.md']);expect(JSON.stringify(state)).not.toContain(studio.directory);
    for(const headers of [{Origin:'null'},{Origin:'https://unrelated.invalid'},{'Sec-Fetch-Site':'cross-site'},{Host:'localhost:'+new URL(studio.origin).port},{'X-Forwarded-Host':'unrelated.invalid'}])expect((await page.request.get(studio.origin+'/studio/state',{headers})).status()).toBe(403);
    expect((await page.request.post(studio.origin+'/studio/state',{data:'overwrite'})).status()).toBe(405);
    expect((await page.request.get(studio.origin+'/%2e%2e/private.md')).status()).toBe(404);
    const response=await page.request.get(studio.origin+'/private/');expect(response.status()).toBe(404);expect(await response.text()).not.toContain('UNSELECTED_PRIVATE_CONTENT');
    const content=source('Safe text\n\n<script>globalThis.__sheetmillInjected=true;fetch("https://sheetmill.invalid/never")</script>\n\n{{< source_shortcode >}}\n\n![External](https://sheetmill.invalid/pixel.png)\n\n```sh\ntouch SHEETMILL_SNIPPET_MUST_NOT_RUN\n```');
    await writeFile(studio.path,content);await current(page,'Safe text');await expect(live(page)).toContainText('touch SHEETMILL_SNIPPET_MUST_NOT_RUN');expect(await page.evaluate(()=>globalThis.__sheetmillInjected)).toBeUndefined();await expect(page.locator('img[src^="https:"]')).toHaveCount(0);
    expect(await readFile(studio.path,'utf8')).toBe(content);await expect(access(join(studio.directory,'SHEETMILL_SNIPPET_MUST_NOT_RUN'))).rejects.toThrow();
    await unlink(studio.path);await symlink('private.md',studio.path);await expect.poll(async()=> (await studio.readState()).sheets[0].status).toBe('invalid');await expect(page.locator('body')).not.toContainText('UNSELECTED_PRIVATE_CONTENT');
  });
});
test.describe('invalid startup',()=>{
  test.use({studioInvalid:true});
  test('starts with diagnostics and recovers without restarting',async({page,studio})=>{
    await page.goto(studio.origin);await expect(page.locator('#studio-status')).toContainText('sample.md:1:1:');
    await writeFile(studio.path,source('Startup repaired'));await expect.poll(async()=> (await studio.readState()).sheets[0].status).toBe('current');
    await page.goto(studio.origin+'/sample/');await current(page,'Startup repaired');
  });
});
