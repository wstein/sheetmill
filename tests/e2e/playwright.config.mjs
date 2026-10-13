// SPDX-License-Identifier: EUPL-1.2
import { defineConfig } from '@playwright/test';
import { fileURLToPath } from 'node:url';
if (process.versions.node !== '24.21.0') throw Error('E2E requires Node 24.21.0');
const evidence = fileURLToPath(new URL('../../.local/e2e/', import.meta.url));
export default defineConfig({
  testDir: '.', testMatch: process.env.SHEETMILL_E2E_CORPUS==='all'?/(?:smill|inert|corpus|assets|copy)\.spec\.mjs$/:/(?:smill|inert|assets|copy)\.spec\.mjs$/,
  outputDir: evidence + 'results', reporter: [['line']], workers: 1, retries: 0,
  timeout: 120_000, expect: { timeout: 10_000 },
  use: { browserName: 'chromium', serviceWorkers: 'block', trace: 'retain-on-failure', screenshot: 'only-on-failure' },
  projects: [
    { name: 'desktop', use: { viewport: { width: 1280, height: 800 } } },
    { name: 'mobile', use: { viewport: { width: 390, height: 844 } } },
    { name: 'no-js', use: { javaScriptEnabled: false, viewport: { width: 1280, height: 800 } } },
  ],
});
