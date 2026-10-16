# Tasks: Sheet freshness

**Input**: [spec.md](spec.md), [plan.md](plan.md), [research.md](research.md),
[data-model.md](data-model.md), [contract](contracts/freshness.md).
**State**: Specification set; implementation starts after the review is READY.
Write each behavioral test first and observe the expected failure before the code.
Commands use `SHEETMILL_HUGO=/opt/bundledock/toolbox/bin/hugo`. No distribution
build, PDF, release, deployment or sibling edit.

## Phase 1: Scope and baseline

- [x] T001 Record the authorization in `AGENTS.md` (one line quoting the 2026-10-16 request) and the status row in `specs/README.md`; draft PR authorized, merge/deployment not. Implementation starts only after this exists. (constitution 7)
- [x] T002 Record base `05eba4d` (the PR base `189e872` is template-equivalent, verified by an empty `git diff 05eba4d 189e872 -- internal/render/hugo/theme`; record both) and baseline results of `go test ./...`, `make check` and `make test-hugo` in `analysis.md`. Save the SHA-256 of every file under `testdata/` and a base export of `smill build navi` for `examples/sheets` (scratch, hashes recorded). Commit the no-field fixture `testdata/hugo/freshness/*.md` and the T009(b) build-and-hash helper (`Render` plus `hugoBuild`, fixed Options recorded in the JSON, pinned Hugo version recorded) behind an explicit `-update-base` flag; run it once on the unchanged templates to write `testdata/hugo/freshness/base-hashes.json` (per-page SHA-256 after normalising `assets/(site|classic|cards).<hash>.css` hrefs). The test is a one-time 017 guard: after 017 merges it is retired or reduced to a no-freshness-markup assertion; regeneration needs a reviewed, deliberate update. (FR-014, SC-003, SC-004)

## Phase 2: Shared validation (US2)

- [x] T003 [US2] Failing table tests in `internal/validate/freshness_test.go`: at least 40 date and 25 version cases (leap days, 1970/9999 bounds, 2026-02-30, `2026-1-5`, Unicode digits, signs, times, numbers, Booleans, null, lists, 64/65 code points, bidi/zero-width/NUL/newline/tab, surrounding spaces, filler-only U+3164/U+2800, invalid UTF-8), absent fields, `cheatsheet` non-mapping, nil locations, positions, determinism, and an unknown `cheatsheet.*` key next to valid freshness (preserved, not rejected, not dropped). Export the accept/reject tables for reuse by T006 and T017. (SC-001, FR-002..005)
- [x] T004 [US2] Implement `validate.Freshness` and `FreshnessSettings` in `internal/validate/freshness.go`; no clock, no reformatting.
- [x] T005 [US2] Failing then passing integration cases in `internal/frontmatter/frontmatter_test.go` (unquoted/quoted YAML date valid, timestamp with time rejected, key position, no partial parse, unknown `cheatsheet.*` key preserved through parse and IR); wire `Freshness` into `validate.Metadata`. Add one CLI test that `smill check` and `smill build navi` fail with `VAL004` on bad input. Add freshness `f.Add` seeds to `FuzzParse` in `internal/frontmatter` and `internal/markdown`. (FR-004, FR-011)
- [x] T006 [US2] Failing then passing direct-IR tests in `internal/render/hugo/freshness_test.go` reusing the T003 tables: invalid values return diagnostics and an empty project; a non-mapping `cheatsheet` is ignored as for `tool`; call `validate.Freshness` beside `validate.Navigation` in `internal/render/hugo/hugo.go`. (US2.5, SC-001)

## Phase 3: Rendering (US1)

- [x] T007 [US1] Failing data tests: `library.json`, `shelves.json`, `sheets/<id>.json`, `tags/<id>.json`, `tagindex.json` carry `updated`/`tool_version` only when set, verbatim; hostile text kept raw in data; shelf order, related, shelf tags, and the search index are identical when dates differ or input order is permuted. (FR-006, FR-009)
- [x] T008 [US1] Add the page fields and populate them from `validate.Freshness` in `internal/render/hugo/hugo.go`; a test asserts the generated `content/<id>/index.md` front matter keys stay `title`, `url`, `sheetid` (freshness only via `data/*.json`) and that `sitemap.xml` has no `<lastmod>`. (FR-009)
- [x] T009 [US1] Failing tests in two layers. (a) Always-run template-text test (`TestShelfCardMarkupInTemplate` pattern) over `home.html`, `page.html` and the partial: exact snippets (including the `.library-fresh` position after the optional `.library-tool` and before `.library-description`, identical with and without `tool`, and including the classic header insertion point after `.sheet-head` and the description, and the index row order `a`, `.index-desc`, `.index-count`, `.index-license`, `.index-fresh` with the freshness span last), `with` guards, `<span>…</span> <span aria-hidden="true">·</span> <span>…</span>` only between two facts, no `title`/`aria-label`/`script`. (b) Real-Hugo test over `editorial`, `classic-compact`, `reference-cards`, home and sheet pages: both/one/none fields, no wrapper when absent, escaped hostile version, `datetime` matches `^[0-9]{4}-[0-9]{2}-[0-9]{2}$` everywhere, tag pages, the tag index and the sidebar partial show none, and the no-field fixture equals the T002 normalised hashes. Named `TestHugoIntegration…` with the existing "SKIPPED, NOT PASSED" message; run with `SHEETMILL_HUGO`. (US1.1-4, SC-003, SC-005)
- [x] T010 [US1] Failing CSS tests (`theme_cards_css_test.go` pattern): freshness rules use palette tokens only, selectors are the scoped forms from the contract, `.classic .index-group li{flex-wrap:wrap}` and `.classic .index-fresh{flex-basis:100%}` exist, `cards.css` resets the basis and centres `.sheet-fresh`, no print `display:none` for them, no `white-space:nowrap` or fixed width on them, `--muted` contrast holds in every palette and mode. This is the CSS-level assertion that replaces the unverifiable "wraps at 320px"; the rendered 320px wrap stays unverified. (FR-008)
- [x] T011 [US1] Add `partials/freshness.html`; use it in `home.html` (card and index row) and `page.html` (both header variants); add CSS to `site.css`, `classic.css` and `cards.css`; keep `TestPerformanceBudget` green. (FR-007, FR-008)
- [x] T012 [US1] Real-Hugo clock test: add a clock-parameterised build helper beside the hard-coded ones; per theme compare (input order A, `--clock 2000-01-01T00:00:00Z`) with (order B, `--clock 2031-06-01T00:00:00Z`); every output file byte-identical. (US1.5, SC-002)

## Phase 4: Import and navi boundaries (US3)

- [x] T013 [US3] Failing tests plus fixtures under `testdata/import/reference/freshness/` (valid, wrong-type, with and without `cards`): archived under `source.original_metadata.cheatsheet`, `IMP112` positioned at the key, no active freshness, no inferred date, cards archive intact, finding order stable. (FR-010)
- [x] T014 [US3] Implement the merged archival in `internal/importer/reference/reference.go` before `validate.Metadata`.
- [x] T015 [P] [US3] Pair regression in `internal/render/navi/navi_test.go`: `.cheat` bytes, entry identities and diagnostics identical with/without the fields; companion keeps them verbatim; version is not a navi tag. Passes without navi code changes. (FR-011, SC-004)

## Phase 5: Contract, goldens, docs, demos

- [x] T016 Needs T005. Add `testdata/reference-compatible/freshness.md` and generate its golden with `go test ./internal/markdown -run TestGolden -update`; inspect semantically and confirm `git status testdata` shows only new files. (FR-014)
- [x] T017 Update `schema/cheatsheet.schema.json` (`updated` pattern, `tool_version` bounds, `$comment` for calendar, invisible and filler rules); add a test that evaluates the schema `pattern`/`maxLength` against the T003 accept/reject table (new patterns stay RE2-compatible; translate any `\uXXXX` to `\x{XXXX}`; count runes for `maxLength`) (the schema may be looser only where its `$comment` says so).
- [ ] T018 Docs: `docs/format/gfm-plus.md` (field table, Freshness section, compatibility note, quote-the-date advice), `docs/diagnostics.md` (`VAL004`, `IMP101`-`IMP112`), `docs/hugo.md`, `docs/imports.md`, `docs/format/navi-profile.md` (not exported, not a tag; parse-time VAL004), `docs/cli.md`, `docs/agent-guide.md` and `llms.txt` (one line: set only from verified evidence). In `docs/hugo.md` and `docs/agent-guide.md` write "sheet freshness (`updated`/`tool_version`)", never bare "freshness" (Studio docs use it for preview staleness); note that the Studio preview shows the facts. (FR-012)
- [ ] T019 Demo data: add `cheatsheet.updated` to all 15 `examples/sheets/*.md` and `tool_version` per [research.md](research.md) (not for `tar`); confirm `smill check` passes, `smill build hugo` into scratch dirs in three themes contains the expected `<time>` facts per sheet, and `smill build navi` `.cheat` output equals the T002 base export. (FR-013, SC-006)

## Phase 6: Verification and handoff

- [ ] T020 Run `gofmt -l`, `go vet ./...`, `go test ./...`, `make check`, `make test-hugo` and `make test-studio` (both with `SHEETMILL_HUGO`; Studio renders through the same templates), `make fuzz` if bounded, and `git diff --check`; report what ran and what did not (browser E2E and the 320px check need Node/a browser and are not run). (SC-001..006)
- [ ] T021 Record review rounds in `analysis.md`, push `codex/017-freshness`, hand over for the draft PR. No merge or deployment.

## Dependencies

T001 -> T002 -> T003 -> T004 -> T005 -> T006. US1: T007 -> T008 -> T009 -> T010 ->
T011 -> T012; T008 needs T004. US3: T013 -> T014; T015 is independent. T016 needs
T005; T017/T018 follow T004; T019 needs T005 and T011. T020/T021 last.
