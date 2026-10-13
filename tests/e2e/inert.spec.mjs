// SPDX-License-Identifier: EUPL-1.2
import { test, expect } from './helpers/fixtures.mjs';

test.use({sourceSuffix:`

<script>globalThis.__sheetmillInjected=true;fetch("https://sheetmill.invalid/never")</script>

{{< source_shortcode >}}

![Remote image](https://sheetmill.invalid/image.png)

\`\`\`sh
touch SHEETMILL_SNIPPET_MUST_NOT_RUN
\`\`\`
`});
test('temporary variants remain inert through CLI, Hugo and browser',async({page,site})=>{
  const errors=[];page.on('pageerror',e=>errors.push(e.message));
  await page.goto(site.origin+'/smill/');
  await expect(page.locator('body')).toContainText('<script>globalThis.__sheetmillInjected=true;fetch("https://sheetmill.invalid/never")</script>');
  await expect(page.locator('body')).toContainText('{{< source_shortcode >}}');
  await expect(page.locator('body')).toContainText('touch SHEETMILL_SNIPPET_MUST_NOT_RUN');
  await expect(page.locator('img[src^="https:"]')).toHaveCount(0);
  expect(await page.evaluate(()=>globalThis.__sheetmillInjected)).toBeUndefined();
  expect(errors).toEqual([]);
  // Source statements cannot select the executable or arguments of a stage.
  for(const r of site.records) {
    if(r.executable.endsWith('/smill'))expect(['check','inspect','build']).toContain(r.args[0]);
    else expect(r.args[0]).toBe('--source');
    expect(r.args).not.toContain('touch');
  }
});
