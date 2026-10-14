# Tasks — 012

- [x] T001 Resolve open decisions for the first slice: GitHub Pages, base path
  `/sheetmill/`, no custom domain, only the independently authored `examples/sheets`.
- [x] T002 Write failing tests for base path validation and sub-path output.
- [x] T003 Implement the base path option; update templates and `app.js`.
- [ ] T004 Write failing tests for the redistribution gate and allow list.
- [ ] T005 Implement the gate and diagnostic; document in docs/cli.md and docs/hugo.md.
- [ ] T006 Complete dedicated CSP lint/header guidance; robots output remains disabled.
- [x] T006a Canonical origin/sitemap implemented (`TestSiteURL*` / sitemap tests in
  internal/render/hugo); no sitemap without an origin.
- [ ] T007 Build twice at root and sub-path; verify bytes, navigation and search.
- [ ] T008 Run required checks and independent review; fix and re-review.
- [ ] T009 Open a reviewable PR (no merge, publication or deployment).

Delivered base-path/manual/demo slices have their own E2E and review evidence in
`docs/e2e-verification.md`. T004–T009 above now refer to the remaining general
redistribution-gate/header work; they do not claim the approved hosted site is
unimplemented or that its earlier approvals cover new content.
