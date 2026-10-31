# SheetMill agent policy

Agent collaboration: see the global rules (`~/.claude/CLAUDE.md`, skill `agent-workflow`); this file contains only repository-specific rules.

SheetMill is a standalone Go reference compiler. The CLI is `smill`.
Use exact product spelling in prose: SheetMill, BundleDock, WorkHarbor, CrewBook.
Sibling repository policies do not govern this repository.

Write all English text in US spelling: documentation, comments, commit messages, pull
request text and identifiers (color, behavior, license, canceled, catalog, labeled,
analyzed, organization). This applies to new and changed text. Do not convert existing
British spellings inside an unrelated change; a separate sweep does that. Keep exact
quotations, testdata, goldens and external names unchanged.

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
- Feature 015 T001–T011 implementation is authorized (user, 2026-10-14):
  target-neutral `.command` cards, shared `cheatsheet.cards.<id>.params`,
  compatibility aliases, core/direct-IR validation and inert import preservation.
  Preserve complete `.cheat` bytes and pinned entry hashes/NAV008. This handoff
  permits isolated implementation, local checks, subagents, independent Sol review
  and a PR. No navi/shell/snippet execution, PDF, merge or deployment.
- Feature 014 is authorized (user, 2026-10-14): per-sheet `navigation.group` and
  `navigation.order` for Hugo library placement, with shared validation, inert
  Reference-import archival and navi-companion preservation (`specs/014-library-navigation`).
  A draft PR is authorized; merge, deployment and reordering the real libraries are not.
  The manifest, group-order file, navigation title and stable IDs stay deferred.
- Feature 016 is authorized (user, 2026-10-14): the opt-in `--theme reference-cards`
  Hugo layout (`specs/016-reference-cards`); editorial stays the default. A reviewed
  PR is authorized; merge and deployment are not.
- Feature 017 is authorized (user, 2026-10-16): optional `cheatsheet.updated` and
  `cheatsheet.tool_version` sheet freshness for Hugo shelf cards and sheet headers
  (`specs/017-freshness`), prompted by "show an `updated` date and the documented tool
  version on library shelf cards and on the sheet header, deterministically". A draft PR
  is authorized; merge and deployment are not.
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
  Extension confirmed by the user on 2026-10-16: the same demo is also published at
  `/sheetmill/demo-compact/` and `/sheetmill/demo-cards/` (paths only; see feature 018).
- Feature 018 is authorized (user, 2026-10-16: "publish the three layouts as separate
  static paths, not a JS switcher"): spec, tests, the Pages verification script, the
  `pages.yml` edit, docs, a topic push and a draft PR (`specs/018-layout-paths`). The
  user confirmed the two new paths on 2026-10-16 (answer "1j"); that covers the paths,
  not their activation. No merge, deployment or workflow run is authorized; merging to
  `main` would activate the workflow and needs separate human approval.
- Feature 019 is authorized (owner, Werner Stein, chat 2026-10-17: "017/018 in v0.1.0, ...
  tag v0.1.0 => CI based release workflow, update CI pipeline"): a tag-triggered release
  workflow (`.github/workflows/release.yml`, `specs/019-release-workflow`). The quote covers
  a CI release workflow in which an owner-pushed tag publishes. There is no standing approval
  (owner decision): every tag and release, including later ones, needs the owner's explicit
  approval each time, and an approval for one tag covers only that tag. The approval is the
  owner's own push of that tag; it is not recorded per version in this file (owner
  clarification 2026-10-21), and this grant is the one standing record of the workflow.
  This authorizes the spec, tests, scripts, workflow, docs, a topic push and a
  draft PR. Merging the workflow PR, creating, moving or deleting any tag (including any
  retag of `v0.1.0`, owner only), running the workflow (dry runs are owner-started
  `workflow_dispatch` only), creating a release, adding secrets, variables or environment
  protection rules, signing, attestations and any other distribution channel remain separate
  human approvals. Re-creating `v0.1.0` is discouraged (the Go module proxy and checksum
  database hold the old `v0.1.0`); the default is a new version such as `v0.1.1`.
- Feature 020 is authorized (owner, Werner Stein, chat 2026-10-20: "ja, Spec 020 entwerfen,
  freigegeben", after the decisions that the Markdown sheet stays the single source and
  that `choices` and `source` are exported as navi `$` lines): `specs/020-navi-variable-sources`.
  The grant covers the specification. Implementation (tests, code, schema, docs, a topic
  push and a draft PR) starts after the owner has merged the specification. Never run
  navi, a shell or an exported command. PDF, merge and deployment are excluded.
- Use `codex/` feature branches (CI runs on `codex/**` pushes). A local-only
  assignment authorizes neither a topic push nor a PR.
- Merge, tag, release, deploy, distribution builds, PDF generation, artifact
  publication and sibling-repository edits require separate human approval.
  Local Hugo site generation and the approved brand exports are authorized.
- Do not execute source snippets, fetch source-linked resources, weaken branch
  protections, add automatic merges or publish artifacts from CI, except the
  authorized Pages deployment and the feature 019 tag-triggered release workflow above.
- Unknown source metadata must be preserved or diagnosed, never silently dropped.

## Pull requests and history

Never push the default branch or tags and never rewrite commits you did not author.
The feature grants above are dated records of what was authorized then (including
"draft PR"); Ready follows the global rules.

## Implementation and verification

Use Go's standard library first; justify and pin dependencies. Keep diagnostic
codes and JSON schema changes deliberate and documented. Use source-positioned
errors, table-driven negative tests, independently authored fixtures, fuzz targets
and byte-for-byte golden JSON.

Run `gofmt`, `go test ./...`, `go vet ./...` and `git diff --check` before review.
Golden changes require semantic inspection; regenerate explicitly with
`go test ./internal/markdown -run TestGolden -update`.

## Light and heavy checks

- Light (every PR and every push to `main` or `codex/**`, workflow `check`): job `core`
  (timeout 10 min: `make fmt-check mod-verify race vet test-assets`, pinned Hugo
  `make test-hugo test-studio`, `make fuzz verify-clean`) and job `build` (timeout 5 min,
  compile-only for linux/amd64, linux/arm64, darwin/arm64). Locally (subset): `make check` runs fmt-check, test, vet and diff-check only.
- Heavy: workflow `web-e2e` (job `web-e2e`, timeout 15 min: Playwright with Chromium and
  pinned Hugo, `make test-harness test-browser test-browser-corpus test-browser-studio`).
  There is no `ci-full` label: it also runs on every PR and push, so there is no separate
  heavy lane at the stack tip. Durations are the workflow timeouts; measured times are UNKNOWN.
- `pages.yml` runs on `main` pushes and manual runs; `release.yml` runs on owner tags and
  manual dry runs (see the grants above).

## Licensing

Original code uses EUPL-1.2 SPDX headers. Keep the full license and provenance.
Implement compatibility independently; do not copy Reference's GPL implementation,
styles or authored sheets. The user selected the Editorial Stack direction with coral/periwinkle colors.
`assets/BRAND.md` governs brand source; the user approved the brand specification/direction and authorized final master
implementation on 2026-10-13. Exact SVG paths and optical variants require review.
Hugo/local generation is authorized; deployment and later renderer gates remain.
