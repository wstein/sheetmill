// SPDX-License-Identifier: EUPL-1.2
import { test as base, expect } from './helpers/fixtures.mjs';
import { OwnedWorkspace, toolEnvironment, registerInterruptCleanup } from './helpers/process.mjs';
import { readSite, compareSites, serveSite } from './helpers/site.mjs';
import { mkdir, writeFile, readFile } from 'node:fs/promises';
import { join } from 'node:path';

const source=(title,category,tags)=>'---\n'+JSON.stringify({title,...(category?{category}:{}),tags})+'\n---\n## Basics\n### Wide {.col-span-3 #wide}\n```sh\necho "SHEETMILL_SNIPPET_MUST_NOT_RUN"\n```\n### Small {#small}\nVisible reference content.\n';
const longTag='longtag'.repeat(30);
const inputs=[['alpha',source('Alpha','Networking',['shared','C++','C#','日本語','<unsafe>/../x',longTag])],['beta',source('Beta','Development',['shared'])],['other',source('Other','',['solo'])]];
const test=base.extend({
  layoutTheme:['editorial',{option:true}],
  placement:[false,{option:true}],
  site:async({tools,layoutTheme,basePath,placement},use)=>{
    const selected=placement?await Promise.all(['zero','essentials','cli','alpha','beta','tie-a','tie-b','net','plain'].map(async id=>[id,await readFile(new URL('../../testdata/hugo/navigation/'+id+'.md',import.meta.url),'utf8')])):inputs;
    const owner=await OwnedWorkspace.create();registerInterruptCleanup(owner);
    const env=toolEnvironment(join(owner.root,'go-cache'));
    const run=(tool,args,extra={})=>owner.run(tool,args,{cwd:owner.root,env,...extra});
    try {
      const input=join(owner.root,'sheets');await mkdir(input);
      for(const [id,bytes] of selected){await writeFile(join(input,id+'.md'),bytes);await run(tools.smill,['check',join(input,id+'.md')]);}
      const sites=[];
      for(const suffix of ['first','second']) {
        const project=join(owner.root,suffix),out=join(project,'public');
        await run(tools.smill,['build','hugo',input,'--out',project,'--base-path',basePath,'--site-url','https://example.invalid'+basePath,'--theme',layoutTheme,'--palette','dracula']);
        await run(tools.hugo,['--source',project,'--destination',out,'--clock','2000-01-01T00:00:00Z','--noBuildLock','--quiet'],{env:{...env,HUGO_CACHEDIR:join(owner.root,'cache-'+suffix)}});
        sites.push(await readSite(out));
      }
      compareSites(...sites);
      const mounted=basePath==='/'?sites[0]:new Map([...sites[0]].map(([name,bytes])=>[basePath.slice(1)+name,bytes]));
      const server=await serveSite(mounted);owner.defer(server.close);
      await use(server);
      for(const [id,bytes] of selected)expect(await readFile(join(input,id+'.md'),'utf8')).toBe(bytes);
    } finally {await owner.close();}
  },
});
const palettes=['graphite-coral','paper-ink','midnight-periwinkle','monochrome','teal-copper','dracula'];
function luminance(hex){const rgb=hex.trim().replace('#','').match(/../g).map(v=>parseInt(v,16)/255).map(v=>v<=.04045?v/12.92:((v+.055)/1.055)**2.4);return rgb[0]*.2126+rgb[1]*.7152+rgb[2]*.0722;}
function contrast(a,b){const [lo,hi]=[luminance(a),luminance(b)].sort((x,y)=>x-y);return(hi+.05)/(lo+.05);}

async function expectFits(page,message) {
  await page.waitForLoadState('load');
  await page.evaluate(()=>document.fonts.ready);
  expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),message).toBe(true);
}

for(const layoutTheme of ['editorial','classic-compact'])for(const basePath of ['/','/sheetmill/'])test.describe(layoutTheme+' navigation at '+basePath,()=>{
  test.use({layoutTheme,basePath});
  test('tag discovery, disclosures and independent palettes',async({page,site},testInfo)=>{
    const errors=[];page.on('pageerror',e=>errors.push(e.message));
    // Exercise navigation readiness with actual stylesheet latency, not a retry.
    if(layoutTheme==='classic-compact' && testInfo.project.name==='mobile')await page.route('**/assets/*.css',async route=>{await new Promise(resolve=>setTimeout(resolve,100));await route.continue();});
    await page.goto(site.origin+basePath+'alpha/');
    const nav=page.getByRole('navigation',{name:'Reference sheets'});
    const active=nav.locator('details[data-library-group="Networking"]');
    await expect(active).toHaveAttribute('open','');
    await expect(nav.getByRole('link',{name:'Alpha',exact:true})).toHaveAttribute('aria-current','page');
    await active.locator('summary').focus();await page.keyboard.press('Enter');await expect(active).not.toHaveAttribute('open','');
    await page.keyboard.press('Space');await expect(active).toHaveAttribute('open','');
    await page.locator('.sheet-tags').getByRole('link',{name:'shared',exact:true}).click();
    await expect(page.getByRole('heading',{name:'Tag: shared',exact:true})).toBeVisible();
    await expect(page.locator('.tag-count')).toHaveText('2 references');
    await expect(page.locator('.tag-results a')).toHaveText(['Alpha','Beta']);
    const sharedURL=page.url();expect(new URL(sharedURL).pathname).toMatch(new RegExp('^'+basePath+'tags/shared-[a-f0-9]{12}/$'));
    expect(await page.locator('link[rel=canonical]').getAttribute('href')).toBe('https://example.invalid'+new URL(sharedURL).pathname);
    await page.reload();await expect(page.locator('.tag-count')).toHaveText('2 references');
    await page.getByRole('link',{name:'← All tags',exact:true}).click();
    await expect(page.getByRole('heading',{name:'Tags',exact:true})).toBeVisible();
    await expect(page.locator('.tag-results a')).toHaveCount(7);
    await page.setViewportSize({width:320,height:844});
    await expectFits(page,'long tag index fits');
    await page.locator('.tag-results').getByRole('link',{name:longTag,exact:true}).click();
    await expect(page.getByRole('heading',{name:'Tag: '+longTag,exact:true})).toBeVisible();
    await expectFits(page,'long tag heading fits');
    await page.getByRole('link',{name:'← All tags',exact:true}).click();

    const unsafe=page.locator('.tag-results').getByRole('link',{name:'<unsafe>/../x',exact:true});
    await unsafe.click();await expect(page.getByRole('heading',{name:'Tag: <unsafe>/../x',exact:true})).toBeVisible();
    await expect(page.locator('.tag-count')).toHaveText('1 reference');
    await page.locator('.tag-results').getByRole('link',{name:'Alpha',exact:true}).click();
    await expect(page.getByRole('heading',{name:'Alpha',exact:true})).toBeVisible();
    await expectFits(page,'long tag chip fits');
    await page.setViewportSize(testInfo.project.use.viewport);
    if(testInfo.project.name==='no-js') {
      await expect(page.locator('#app-palette')).toBeHidden();
      await page.emulateMedia({colorScheme:'dark'});
      expect(await page.locator('body').evaluate(el=>getComputedStyle(el).backgroundColor)).toBe('rgb(40, 42, 54)');
    } else {
      const group=nav.locator('details[data-library-group="Development"]');
      await group.locator('summary').click();await expect(group).toHaveAttribute('open','');
      await group.locator('summary').click();await expect(group).not.toHaveAttribute('open','');
      await page.reload();await expect(group).not.toHaveAttribute('open','');
      // A current page must reveal its group even after a stored collapse.
      await page.goto(site.origin+basePath+'beta/');await expect(group).toHaveAttribute('open','');
      const palette=page.getByRole('combobox',{name:'Color palette'});
      await expect(palette.locator('option')).toHaveCount(6);
      for(const name of palettes) {
        await palette.selectOption(name);
        for(const mode of ['dark','light']) {
          if(await page.locator('html').getAttribute('data-theme')!==mode)await page.locator('#app-theme').click();
          const values=await page.evaluate(()=>{const s=getComputedStyle(document.documentElement);return Object.fromEntries(['bg','surface','code','text','muted','link','accent'].map(k=>[k,s.getPropertyValue('--'+k).trim()]));});
          for(const fg of ['text','muted','link','accent'])for(const bg of ['bg','surface','code'])expect(contrast(values[fg],values[bg]),name+' '+mode+' '+fg+'/'+bg).toBeGreaterThanOrEqual(4.5);
          if(testInfo.project.name==='desktop' && basePath==='/sheetmill/')await page.screenshot({path:testInfo.outputPath(name+'-'+mode+'.png'),fullPage:true});
        }
        await page.reload();await expect(page.locator('html')).toHaveAttribute('data-palette',name);
      }
      if(layoutTheme==='classic-compact')await page.locator('#app-density').click();
    }
    for(const width of [320,900]) {
      await page.setViewportSize({width,height:844});
      await expectFits(page,'fits '+width);
    }
    expect(errors).toEqual([]);
  });
});

for(const layoutTheme of ['editorial','classic-compact'])for(const basePath of ['/','/sheetmill/'])test.describe(layoutTheme+' placement at '+basePath,()=>{
  test.use({layoutTheme,basePath,placement:true});
  test('authored placement agrees in home and sidebar',async({page,site})=>{
    await page.goto(site.origin+basePath);
    await expectFits(page,'placement home fits');
    const nav=page.getByRole('navigation',{name:'Reference sheets'});
    const labels=['Networking','SheetMill','reference','Other references'];
    const titles=[['Net'],['Zero','Essentials','CLI','alpha','Beta'],['Tie','Tie'],['Plain']];
    await expect(nav.locator('summary')).toHaveText(['Networking 1','SheetMill 5','reference 2','Other references 1']);
    const shelves=page.locator('.shelf, .index-group');
    await expect(shelves).toHaveCount(labels.length);
    for(let i=0;i<labels.length;i++) {
      const group=nav.locator('details').nth(i);
      await expect(shelves.nth(i).getByRole('heading',{level:2})).toHaveText(labels[i]);
      const links=group.locator('a');
      await expect(links).toHaveText(titles[i]);
      const homeLinks=shelves.nth(i).locator('a');
      expect(await homeLinks.evaluateAll(els=>els.map(el=>el.getAttribute('href')))).toEqual(await links.evaluateAll(els=>els.map(el=>el.getAttribute('href'))));
    }
    const group=nav.locator('details[data-library-group="SheetMill"]');
    if(!await group.evaluate(el=>el.open))await group.locator('summary').click();
    await group.getByRole('link',{name:'CLI',exact:true}).click();
    await expect(page.getByRole('heading',{level:1,name:'CLI',exact:true})).toBeVisible();
    await expect(group).toHaveAttribute('open','');
    await expect(group.getByRole('link',{name:'CLI',exact:true})).toHaveAttribute('aria-current','page');
    await expect(group.locator('a')).toHaveText(titles[1]);
    await group.locator('summary').focus();await page.keyboard.press('Enter');
    await expect(group).not.toHaveAttribute('open','');
    await page.keyboard.press('Space');await expect(group).toHaveAttribute('open','');
    await expectFits(page,'placement sheet fits');
  });
});
