// SPDX-License-Identifier: EUPL-1.2
import base from './playwright.config.mjs';
import { fileURLToPath } from 'node:url';
export default {...base,outputDir:fileURLToPath(new URL('../../.local/e2e/performance/run-'+Date.now()+'-'+process.pid+'/',import.meta.url)),testMatch:/performance\.spec\.mjs$/,timeout:900_000,retries:0,workers:1,
  use:{...base.use,trace:'off',screenshot:'off',video:'off'},
  projects:base.projects.filter(p=>p.name==='desktop')};
