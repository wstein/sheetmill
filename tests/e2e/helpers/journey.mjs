// SPDX-License-Identifier: EUPL-1.2
import { expect } from './fixtures.mjs';

export async function readerJourney({ page, context, site, corpus, corpusCases }, testInfo) {
  for(const stage of site.toolRecords) expect(stage.environment).toEqual({GOPROXY:'off',GOSUMDB:'off',GOTOOLCHAIN:'local'});
  const errors=[];page.on('pageerror',error=>errors.push(error.message));
  await page.goto(site.origin);
  await expect(page).toHaveTitle('SheetMill References');
  const library=page.getByRole('navigation',{name:'Reference sheets'});
  await expect(library.getByRole('link')).toHaveCount(corpusCases.length);
  await expect(library.getByRole('link',{name:corpus.title,exact:true})).toBeVisible();
  if(testInfo.project.name==='no-js') {
    await library.getByRole('link',{name:corpus.title,exact:true}).click();
  } else {
    const search=page.getByRole('searchbox',{name:'Find a reference'});
    await expect(search).toBeEnabled();await search.fill(corpus.search.query);
    const result=page.locator('#app-results').getByRole('link',{name:corpus.search.label,exact:true});
    await expect(result).toBeVisible();await search.press('Tab');await expect(result).toBeFocused();await result.press('Enter');
    await expect(page).toHaveURL(site.origin+`/${corpus.id}/#ref-${corpus.search.card}`);
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
  const permalink=page.getByRole('link',{name:corpus.permalink.label,exact:true});
  await permalink.click();await expect(page).toHaveURL(site.origin+`/${corpus.id}/#ref-${corpus.permalink.card}`);
  await expect(page.locator('.provenance')).toContainText(corpus.attribution.license);
  await expect(page.locator('.provenance')).toContainText(corpus.attribution.author);
  const sourceDetails = page.locator('.source-details');
  await expect(sourceDetails).not.toHaveAttribute('open', '');
  await expect(sourceDetails.locator('.source-metadata')).not.toBeVisible();
  await sourceDetails.locator('summary').click();
  await expect(sourceDetails.locator('.source-metadata')).toBeVisible();
  await expect(sourceDetails.locator('.source-metadata')).toContainText(corpus.attribution.license);
  await expect(sourceDetails.locator('dt')).toContainText(['Attribution:', 'License:']);
  await expect(sourceDetails.locator('pre')).toHaveCount(0);
  await sourceDetails.locator('summary').click();
  await expect(sourceDetails.locator('.source-metadata')).not.toBeVisible();
  expect(await page.evaluate(()=>document.documentElement.scrollWidth<=document.documentElement.clientWidth)).toBe(true);
  for(const link of corpus.crossLinks) {
    await page.getByRole('link',{name:link.label,exact:true}).click();
    await expect(page).toHaveURL(site.origin+`/${link.sheet}/#ref-${link.card}`);
    await expect(page.locator('#ref-'+link.card)).toBeVisible();
    await expect(page.getByRole('heading',{name:corpusCases.find(c=>c.id===link.sheet).title,exact:true})).toBeVisible();
    await page.goto(site.origin+`/${corpus.id}/`);
  }
  expect(errors).toEqual([]);
}
