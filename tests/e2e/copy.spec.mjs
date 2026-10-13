// SPDX-License-Identifier: EUPL-1.2
import { test,expect } from './helpers/fixtures.mjs';
const text='long_token_'.repeat(40)+'\nsecond line\tvalue\n';
test.use({sourceSuffix:'\n## Copy layout fixture\n### Long code {#copy-layout}\n```text\n'+text+'```\n'});

test('compact copy control preserves layout, keyboard copying and failure feedback',async({page,context,site},testInfo)=>{
  await page.goto(site.origin+'/smill/#ref-copy-layout');
  const block=page.locator('#ref-copy-layout pre'),code=block.locator('code');
  await expect(code).toHaveText(text,{useInnerText:false});
  if(testInfo.project.name==='no-js'){
    await expect(block.locator('.copy')).toHaveCount(0);return;
  }
  const button=block.getByRole('button',{name:'Copy code',exact:true});await expect(button).toBeVisible();
  const layout=await block.evaluate(el=>{
    const button=el.querySelector('.copy').getBoundingClientRect(),range=document.createRange();range.selectNodeContents(el.querySelector('code'));
    const rects=[...range.getClientRects()].filter(r=>r.width>0&&r.height>0);
    return {paddingTop:parseFloat(getComputedStyle(el).paddingTop),firstLineOffset:rects[0].top-el.getBoundingClientRect().top,width:button.width,
      overlaps:rects.some(r=>r.left<button.right&&r.right>button.left&&r.top<button.bottom&&r.bottom>button.top)};
  });
  expect(layout.paddingTop).toBe(14);expect(layout.firstLineOffset).toBeLessThan(25);expect(layout.overlaps).toBe(false);
  expect(layout.width).toBeGreaterThanOrEqual(24);expect(layout.width).toBeLessThanOrEqual(36);
  await context.grantPermissions(['clipboard-read','clipboard-write'],{origin:site.origin});
  await button.focus();await page.keyboard.press('Enter');
  await expect(block.getByRole('button',{name:'Copied',exact:true})).toBeVisible();
  await expect(block.getByRole('status')).toHaveText('Code copied');
  await expect.poll(()=>page.evaluate(()=>navigator.clipboard.readText())).toBe(text);
  await expect(block.getByRole('button',{name:'Copy code',exact:true})).toBeVisible();
  await page.evaluate(()=>{navigator.clipboard.writeText=async()=>{throw new Error('Clipboard denied')}});
  await button.click();const failed=block.getByRole('button',{name:'Select code to copy',exact:true});await expect(failed).toBeVisible();
  await expect(block.getByRole('status')).toHaveText('Copy failed. Select code to copy.');
  expect((await failed.boundingBox()).width).toBe(layout.width);await expect(code).toHaveText(text,{useInnerText:false});
  expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
});
