# SheetMill agent policy

SheetMill is a standalone Go reference compiler. The CLI is `smill`.
Use exact product spelling in prose: SheetMill, BundleDock, WorkHarbor, CrewBook.
Sibling repository policies do not govern this repository.

## Scope and authorization

- Routine local source/document edits, dependency inspection, formatting, tests,
  fuzzing, vet and source/security checks are allowed within the requested work.
- The current scope includes features 001–004: core, Reference import, finalized
  brand assets and a Hugo renderer with local site generation (authorized by the
  user on 2026-10-13). PDF/Typst renderers remain out of scope.
- Feature 006 is authorized (user, 2026-10-14): `smill build navi` text generation
  of `.cheat` files for bash, zsh, fish and nushell profiles. Never run navi, a shell
  or an exported command; publishing generated cheats and distribution stay gated.
  T009 deterministic metadata sidecars are authorized on 2026-10-14; this increment
  permits implementation, local verification and a reviewable PR only. No PDF or
  deployment.
- Feature 015 command cards is authorized for documentation only (user,
  2026-10-14): define target-neutral `.command` cards and
  `cheatsheet.cards.<id>.params`, migration from 006 syntax and unchanged navi
  identity/output requirements. This permits spec/format/006 updates, local
  documentation checks, independent Sol review and a docs PR. Core/parser,
  renderer, schema and migration-tool implementation require a later handoff.
  No navi execution, PDF, merge or deployment.
- Feature 014 is authorized (user, 2026-10-14): per-sheet `navigation.group` and
  `navigation.order` for Hugo library placement, with shared validation, inert
  Reference-import archival and navi-companion preservation (`specs/014-library-navigation`).
  A draft PR is authorized; merge, deployment and reordering the real libraries are not.
  The manifest, group-order file, navigation title and stable IDs stay deferred.
- Feature 007 MVP T001–T018 and applicable final checks are authorized on
  2026-10-13: one smill sheet, temporary verification executable, Hugo and local
  browser E2E.
- Feature 007 US2 T042–T046 and applicable final checks are authorized on
  2026-10-13: independently authored BundleDock/WorkHarbor references through
  the same harness. Sibling commands remain content; no sibling runtime or edits.
- Feature 007 T019–T041 and applicable final checks are authorized on
  2026-10-13: read-only local `smill serve`, truthful freshness/diagnostics,
  scoped source access and authenticated loopback HTTP. Browser editing, PDF,
  navi, playgrounds and deployment are excluded from this increment.
- Feature 012 is partly authorized (user, 2026-10-13; extended 2026-10-14): the base
  path option and GitHub Pages deployment, from `main` pushes and manual runs, of
  the SheetMill manual built from the user-facing `docs/` pages at
  https://wstein.github.io/sheetmill/, the independently authored `examples/sheets`
  demo at `/sheetmill/demo/`, and `llms.txt` plus the agent guide as plain files.
  The redistribution gate and other sheets or hosts remain separately gated.
- Use `codex/` feature branches and Conventional Commits. Topic pushes and draft
  PRs belong to an authorized implementation handoff; otherwise ask before publishing.
- Merge, tag, release, deploy, distribution builds, PDF generation, artifact
  publication and sibling-repository edits require separate human approval.
  Local Hugo site generation and the approved brand exports are authorized.
- Do not execute source snippets, fetch source-linked resources, weaken branch
  protections, add automatic merges or publish artifacts from CI, except the
  authorized Pages deployment above.
- Unknown source metadata must be preserved or diagnosed, never silently dropped.

## Implementation and verification

Use Go's standard library first; justify and pin dependencies. Keep diagnostic
codes and JSON schema changes deliberate and documented. Use source-positioned
errors, table-driven negative tests, independently authored fixtures, fuzz targets
and byte-for-byte golden JSON. Write failing behavior tests before implementation.

Run `gofmt`, `go test ./...`, `go vet ./...` and `git diff --check` before review.
Golden changes require semantic inspection; regenerate explicitly with
`go test ./internal/markdown -run TestGolden -update`.
Report commands actually run and remaining gaps. Independent human or separately
assigned reviewer approval is expected before merge. An author's own checks do
not constitute independent review. Delegation is not required by this policy.

## Licensing

Original code uses EUPL-1.2 SPDX headers. Keep the full license and provenance.
Implement compatibility independently; do not copy Reference's GPL implementation,
styles or authored sheets. The user selected the Editorial Stack direction with coral/periwinkle colors.
`assets/BRAND.md` governs brand source; the user approved the brand specification/direction and authorized final master
implementation on 2026-10-13. Exact SVG paths and optical variants require review.
Hugo/local generation is authorized; deployment and later renderer gates remain.
