// SPDX-License-Identifier: EUPL-1.2
import { test, expect } from './helpers/fixtures.mjs';

test('one smill sheet through the complete reader journey', async ({ page, context, site, corpus }, testInfo) => {
  const errors=[];page.on('pageerror',error=>errors.push(error.message));
  await page.goto(site.origin);
  await expect(page).toHaveTitle('SheetMill References');
  const library=page.getByRole('navigation',{name:'Reference sheets'});
  await expect(library.getByRole('link')).toHaveCount(1);
  await expect(library.getByRole('link',{name:corpus.title,exact:true})).toBeVisible();
  if(testInfo.project.name==='no-js') {
    await library.getByRole('link',{name:corpus.title,exact:true}).click();
  } else {
    const search=page.getByRole('searchbox',{name:'Find a reference'});
    await expect(search).toBeEnabled();await search.fill(corpus.search.query);
    const result=page.locator('#app-results').getByRole('link',{name:corpus.search.label,exact:true});
    await expect(result).toBeVisible();await search.press('Tab');await expect(result).toBeFocused();await result.press('Enter');
    await expect(page).toHaveURL(site.origin+`/smill/#ref-${corpus.search.card}`);
  }
  await expect(page.getByRole('heading',{name:corpus.title,exact:true})).toBeVisible();
  const card=page.locator('#ref-'+corpus.copy.card);
  await expect(card.locator('pre > code')).toHaveText(corpus.copy.text,{useInnerText:false});
  // textContent is also compared without normalization to preserve exact newlines.
  expect(await card.locator('pre > code').textContent()).toBe(corpus.copy.text);
  for(const id of corpus.cards) await expect(page.locator('#ref-'+id)).toHaveCount(1);
  if(testInfo.project.name!=='no-js') {
    await context.grantPermissions(['clipboard-read','clipboard-write'],{origin:site.origin});
    await page.evaluate(()=>navigator.clipboard.writeText('sheetmill-e2e-sentinel'));
    await card.getByRole('button',{name:'Copy code',exact:true}).click();
    await expect.poll(()=>page.evaluate(()=>navigator.clipboard.readText())).toBe(corpus.copy.text);
    for(const mode of ['dark','light']) {
      await page.getByRole('button',{name:mode==='dark'?'Dark mode':'Light mode',exact:true}).click();
      await expect(page.locator('html')).toHaveAttribute('data-theme',mode);
      await expect.poll(()=>page.locator('body').evaluate(el=>getComputedStyle(el).backgroundColor)).toBe(mode==='dark'?'rgb(44, 49, 64)':'rgb(247, 248, 252)');
      await page.reload();await expect(page.locator('html')).toHaveAttribute('data-theme',mode);
    }
  } else {
    await expect(page.getByRole('button',{name:/mode/})).toHaveCount(0);
    await expect(page.getByRole('searchbox',{name:'Find a reference'})).toBeDisabled();
  }
  const permalink=page.getByRole('link',{name:'Permalink to Exit status',exact:true});
  await permalink.click();await expect(page).toHaveURL(site.origin+'/smill/#ref-exit-status');
  await expect(page.locator('.provenance')).toContainText(corpus.attribution.license);
  await expect(page.locator('.provenance')).toContainText(corpus.attribution.author);
  expect(await page.evaluate(()=>document.documentElement.scrollWidth<=document.documentElement.clientWidth)).toBe(true);
  expect(errors).toEqual([]);
});
