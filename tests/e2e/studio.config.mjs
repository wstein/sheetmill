// SPDX-License-Identifier: EUPL-1.2
import base from './playwright.config.mjs';
export default {...base,testMatch:/studio\.spec\.mjs$/,use:{...base.use,trace:'off'},projects:base.projects.filter(p=>p.name!=='no-js')};
