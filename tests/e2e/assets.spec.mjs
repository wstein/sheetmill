// SPDX-License-Identifier: EUPL-1.2
import { createHash } from 'node:crypto';
import { test, expect } from './helpers/fixtures.mjs';

for (const basePath of ['/', '/sheetmill/']) {
  test.describe('fingerprinted assets at '+basePath, () => {
    test.use({basePath});
    test('browser loads content-addressed CSS and JavaScript', async ({page, site}, testInfo) => {
      const responses=[], errors=[];
      page.on('response', response => {
        if (['stylesheet','script','font'].includes(response.request().resourceType())) responses.push(response);
      });
      page.on('pageerror', error => errors.push(error.message));
      await page.goto(site.origin+basePath);
      await page.evaluate(() => document.fonts.ready);
      const paths=[];
      for (const [selector, attribute, stem, extension, contentType] of [
        ['link[rel="stylesheet"]','href','site','css','text/css'],
        ['script[src]','src','app','js','javascript'],
      ]) {
        const url=new URL(await page.locator(selector).getAttribute(attribute),site.origin);
        expect(url.origin).toBe(site.origin);
        expect(url.search).toBe('');
        expect(url.pathname.startsWith(basePath+'assets/')).toBe(true);
        const match=url.pathname.slice(basePath.length).match(new RegExp('^assets/'+stem+'\\.([a-f0-9]{64})\\.'+extension+'$'));
        expect(match,'asset URL contains its SHA-256 digest').not.toBeNull();
        const response=await page.request.get(url.href);
        expect(response.status()).toBe(200);
        expect(response.headers()['content-type']).toContain(contentType);
        expect(createHash('sha256').update(await response.body()).digest('hex')).toBe(match[1]);
        paths.push(url.pathname);
      }
      // CSS loads without JS; script execution remains a progressive enhancement.
      expect(responses.some(r=>new URL(r.url()).pathname===paths[0] && r.status()===200)).toBe(true);
      if(testInfo.project.name!=='no-js') {
        expect(responses.some(r=>new URL(r.url()).pathname===paths[1] && r.status()===200)).toBe(true);
        await expect(page.getByRole('searchbox',{name:'Find a reference'})).toBeEnabled();
      }
      expect(responses.some(r=>r.request().resourceType()==='font' && r.status()===200)).toBe(true);
      for(const response of responses) {
        expect(response.status()).toBe(200);
        const digest=new URL(response.url()).pathname.match(/\.([a-f0-9]{64})\.(?:css|js)$/)?.[1];
        if(digest) expect(createHash('sha256').update(await response.body()).digest('hex')).toBe(digest);
      }
      // Home and sheet pages share the same fingerprinted URLs.
      await page.goto(site.origin+basePath+'smill/');
      expect(new URL(await page.locator('link[rel="stylesheet"]').getAttribute('href'),site.origin).pathname).toBe(paths[0]);
      expect(new URL(await page.locator('script[src]').getAttribute('src'),site.origin).pathname).toBe(paths[1]);
      expect(responses.some(r=>/\/assets\/(site\.css|app\.js)$/.test(new URL(r.url()).pathname))).toBe(false);
      expect(errors).toEqual([]);
    });
  });
}
