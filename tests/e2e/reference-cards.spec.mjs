// SPDX-License-Identifier: EUPL-1.2
import {test as base,expect} from './helpers/fixtures.mjs';
import {OwnedWorkspace,toolEnvironment,registerInterruptCleanup} from './helpers/process.mjs';
import {readSite,compareSites,serveSite} from './helpers/site.mjs';
import {mkdir,writeFile,readFile} from 'node:fs/promises';
import {join} from 'node:path';

const command='printf "%s" "SHEETMILL_SNIPPET_MUST_NOT_RUN"\n';
const longTitle='LongUnbrokenBadge'.repeat(12);
const source='---\n'+JSON.stringify({title:'Theme reference',description:'Command examples and readable reference tiles.',category:'Tools',tags:['theme'],navigation:{group:'SheetMill',order:20}})+'\n---\n## Examples\n### Copy literal {#copy}\nDescription before a shell example.\n```bash\n'+command+'```\n### '+longTitle+' {#long}\nA long badge stays above its content.\n### Wide {#wide .col-span-3}\n| Option | Meaning |\n| --- | --- |\n| `'+('longoption'.repeat(50))+'` | A scrollable table |\n### Two columns {#two .col-span-2}\nDescription before plain text.\n```text\nPlain text has no shell prompt.\n```\n';
const other='---\ntitle: First reference\nnavigation: {group: SheetMill, order: 0}\n---\n## Basics\n### Note {#note}\nInert reference text.\n';
const test=base.extend({
 site:async({tools,basePath},use)=>{
  const owner=await OwnedWorkspace.create();registerInterruptCleanup(owner);
  const env=toolEnvironment(join(owner.root,'go-cache'));
  const run=(tool,args,extra={})=>owner.run(tool,args,{cwd:owner.root,env,...extra});
  try {
   const input=join(owner.root,'sheets');await mkdir(input);
   await writeFile(join(input,'sample.md'),source);await writeFile(join(input,'other.md'),other);
   const sites=[];
   for(const suffix of ['first','second']) {
    const project=join(owner.root,suffix),out=join(project,'public');
    await run(tools.smill,['build','hugo',input,'--out',project,'--base-path',basePath,'--theme','reference-cards','--description','Reference card library.']);
    await run(tools.hugo,['--source',project,'--destination',out,'--clock','2000-01-01T00:00:00Z','--noBuildLock','--quiet'],{env:{...env,HUGO_CACHEDIR:join(owner.root,'cache-'+suffix)}});
    sites.push(await readSite(out));
   }
   compareSites(...sites);
   const mounted=basePath==='/'?sites[0]:new Map([...sites[0]].map(([name,bytes])=>[basePath.slice(1)+name,bytes]));
   const server=await serveSite(mounted);owner.defer(server.close);await use(server);
   expect(await readFile(join(input,'sample.md'),'utf8')).toBe(source);
  }finally{await owner.close();}
 },
});
function contrast(a,b){const luminance=rgb=>{const c=rgb.match(/[0-9.]+/g).slice(0,3).map(Number).map(v=>{v/=255;return v<=.04045?v/12.92:((v+.055)/1.055)**2.4});return c[0]*.2126+c[1]*.7152+c[2]*.0722};const [lo,hi]=[luminance(a),luminance(b)].sort((x,y)=>x-y);return(hi+.05)/(lo+.05);}
async function loaded(page){await page.waitForLoadState('load');await page.evaluate(()=>document.fonts.ready);}
async function centered(page){
 const delta=await page.locator('.sheet-description').evaluate(el=>{const a=el.getBoundingClientRect(),b=document.querySelector('#app-main').getBoundingClientRect();return Math.abs((a.left+a.right)/2-(b.left+b.right)/2)});
 expect(delta,'description centered with heading').toBeLessThanOrEqual(1);
}
async function fits(page){await loaded(page);expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),'no document overflow').toBe(true);}

for(const basePath of ['/','/sheetmill/'])test.describe('reference-cards at '+basePath,()=>{
 test.use({basePath});
 test('native navigation, tiles, palette, density, copy and print media',async({page,context,site},testInfo)=>{
  await page.goto(site.origin+basePath);await loaded(page);await centered(page);
  await expect(page.locator('body')).toHaveClass('classic reference-cards');
  await expect(page.locator('.library-sidebar')).toHaveCount(0);
  const styles=await page.locator('link[rel=stylesheet]').evaluateAll(els=>els.map(el=>el.getAttribute('href')));
  expect(styles.some(href=>new RegExp('^'+basePath+'assets/cards\\.[a-f0-9]{64}\\.css$').test(href))).toBe(true);
  const browse=page.locator('.browse-panel');await expect(browse).not.toHaveAttribute('open','');
  await browse.locator('summary').first().focus();await page.keyboard.press('Enter');await expect(browse).toHaveAttribute('open','');
  const group=browse.locator('details[data-library-group="SheetMill"]');
  await expect(group.locator('a')).toHaveText(['First reference','Theme reference']);
  await group.getByRole('link',{name:'Theme reference',exact:true}).click();
  await loaded(page);await centered(page);
  await expect(group).toHaveAttribute('open','');
  await browse.locator('summary').first().click();
  await expect(group.getByRole('link',{name:'Theme reference',exact:true})).toHaveAttribute('aria-current','page');
  await expect(group.getByRole('link',{name:'Theme reference',exact:true})).toBeVisible();
  const block=page.locator('#ref-copy pre'),code=block.locator('code');
  await expect(code).toHaveText(command,{useInnerText:false});
  expect(await code.evaluate(el=>getComputedStyle(el,'::before').content)).toBe('"$ "');
  expect(await code.evaluate(el=>getComputedStyle(el).backgroundColor),'block code has no inline-code tint').toBe('rgba(0, 0, 0, 0)');
  expect(await page.locator('#ref-two pre code').evaluate(el=>getComputedStyle(el,'::before').content)).toBe('none');
  expect(await page.locator('#ref-wide td code').evaluate(el=>getComputedStyle(el).backgroundColor)).not.toBe('rgba(0, 0, 0, 0)');
  if(testInfo.project.name!=='no-js') {
   await context.grantPermissions(['clipboard-read','clipboard-write'],{origin:site.origin});
   await block.getByRole('button',{name:'Copy code',exact:true}).click();
   await expect.poll(()=>page.evaluate(()=>navigator.clipboard.readText())).toBe(command);
   for(const palette of ['graphite-coral','paper-ink','midnight-periwinkle','monochrome','teal-copper','dracula']) {
    await page.getByRole('combobox',{name:'Color palette'}).selectOption(palette);
    for(const theme of ['light','dark']) {
     if(await page.locator('html').getAttribute('data-theme')!==theme)await page.locator('#app-theme').click();
     const colors=await page.locator('#ref-copy h3').evaluate(el=>({text:getComputedStyle(el).color,background:getComputedStyle(el).backgroundColor}));
     expect(contrast(colors.text,colors.background),palette+' '+theme+' badge contrast').toBeGreaterThanOrEqual(4.5);
    }
   }
  }
  if(testInfo.project.name==='no-js') {
   await page.emulateMedia({colorScheme:'dark'});
   const colors=await page.locator('#ref-copy h3').evaluate(el=>({text:getComputedStyle(el).color,background:getComputedStyle(el).backgroundColor}));
   expect(contrast(colors.text,colors.background)).toBeGreaterThanOrEqual(4.5);
  }
  for(const density of (testInfo.project.name==='no-js'?['compact']:['compact','comfortable'])) {
   if(density==='comfortable') {
    await page.locator('#app-density').click();
    await expect(page.locator('body')).toHaveClass(/comfortable/);
   } else await expect(page.locator('body')).not.toHaveClass(/comfortable/);
   for(const width of [320,390,640,650,700,760,900,1100,1280]) {
    await page.setViewportSize({width,height:844});await fits(page);
    const badge=await page.locator('#ref-long').evaluate(el=>{const a=el.querySelector('h3').getBoundingClientRect(),b=el.querySelector('p').getBoundingClientRect();return {bottom:a.bottom,nextTop:b.top}});
    expect(badge.bottom).toBeLessThanOrEqual(badge.nextTop);
    for(const id of ['wide','two'])expect(await page.locator('#ref-'+id).evaluate(el=>el.getBoundingClientRect().right<=document.querySelector('#app-main').getBoundingClientRect().right)).toBe(true);
   }
  }
  await page.setViewportSize(testInfo.project.use.viewport);await fits(page);
  await page.locator('#ref-wide table').scrollIntoViewIfNeeded();
  expect(await page.locator('#ref-wide table').evaluate(el=>el.scrollWidth>el.clientWidth),'wide table scrolls inside tile').toBe(true);
  expect(await page.locator('#ref-wide table').evaluate(el=>{el.scrollLeft=el.scrollWidth;return el.scrollLeft})).toBeGreaterThan(0);
  await expect(page.getByRole('columnheader',{name:'Meaning',exact:true})).toBeInViewport();
  expect(await page.locator('#ref-wide th').nth(1).evaluate(el=>{const r=document.createRange();r.selectNodeContents(el);return r.getClientRects().length}),'short table header stays readable').toBe(1);
  await page.locator('#ref-wide table').evaluate(el=>{el.scrollLeft=0});
  if(testInfo.project.name!=='no-js')await page.evaluate(()=>new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve))));
  await page.screenshot({path:testInfo.outputPath('reference-cards-screen.png'),fullPage:true});
  await page.emulateMedia({media:'print'});await loaded(page);
  await expect(browse).toBeHidden();
  expect(await page.locator('#ref-wide table').evaluate(el=>el.scrollWidth<=el.clientWidth),'print table wraps within tile').toBe(true);
  expect(await page.locator('#ref-copy').evaluate(el=>getComputedStyle(el).boxShadow)).toBe('none');
  expect(await page.locator('#ref-copy').evaluate(el=>getComputedStyle(el).borderTopStyle)).toBe('solid');
  expect(await page.locator('#ref-copy h3').evaluate(el=>getComputedStyle(el).color)).toBe('rgb(0, 0, 0)');
  expect(await page.locator('#ref-copy p').evaluate(el=>getComputedStyle(el).backgroundColor)).toMatch(/^(rgba\(0, 0, 0, 0\)|rgb\(255, 255, 255\))$/);
  await page.screenshot({path:testInfo.outputPath('reference-cards-print-media.png'),fullPage:true});
 });
});
