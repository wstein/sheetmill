# Tasks: Sheet freshness

**Input**: [spec.md](spec.md), [plan.md](plan.md), [research.md](research.md),
[data-model.md](data-model.md), [contract](contracts/freshness.md).
**State**: Specification set; implementation starts after the review is READY.
Write each behavioral test first and observe the expected failure before the code.
Commands use `SHEETMILL_HUGO=/opt/bundledock/toolbox/bin/hugo`. No distribution
build, PDF, release, deployment or sibling edit.

## Phase 1: Scope and baseline

- [ ] T001 Record the authorization in `AGENTS.md` (one line) and the status row in `specs/README.md`; draft PR authorized, merge/deployment not. (FR-012)
- [ ] T002 Record the base `05eba4d` and the baseline results of `go test ./...`, `make check` and `make test-hugo` in `specs/017-freshness/analysis.md`; save the SHA-256 of every `testdata/**/*.json` golden for the later byte-identity claim. (FR-014, SC-004)

## Phase 2: Shared validation (US2)

- [ ] T003 [US2] Failing table tests in `internal/validate/freshness_test.go`: at least 40 date and 25 version cases from the spec (leap days, 1970/9999 bounds, 2026-02-30, `2026-1-5`, Unicode digits, signs, times, numbers, Booleans, null, lists, 64/65 code points, bidi/zero-width/NUL/newline/tab, surrounding spaces, invalid UTF-8), absent fields, `cheatsheet` non-mapping, nil locations, positions, determinism. (SC-001, FR-002..005)
- [ ] T004 [US2] Implement `validate.Freshness` and `FreshnessSettings` in `internal/validate/freshness.go`; no clock, no reformatting.
- [ ] T005 [US2] Failing then passing integration cases in `internal/frontmatter/frontmatter_test.go` (unquoted/quoted YAML date both valid, timestamp with time rejected, key position, no partial parse); wire `Freshness` into `validate.Metadata`. (FR-004)
- [ ] T006 [US2] Failing then passing direct-IR tests in `internal/render/hugo/freshness_test.go`: invalid values return diagnostics and an empty project; call `validate.Freshness` beside `validate.Navigation` in `internal/render/hugo/hugo.go`. (US2.5)

## Phase 3: Rendering (US1)

- [ ] T007 [US1] Failing data tests in `internal/render/hugo/freshness_test.go`: `library.json`, `shelves.json`, `sheets.json` carry `updated`/`tool_version` only when set, verbatim; hostile text kept raw in data; ordering, related and shelf tags unchanged when dates differ or are permuted. (FR-006, FR-009)
- [ ] T008 [US1] Add the page fields and populate them from `validate.Freshness` in `internal/render/hugo/hugo.go`.
- [ ] T009 [US1] Failing markup tests, table-driven over `editorial`, `classic-compact`, `reference-cards`, home and sheet pages: exact snippets for both/one/none fields, no wrapper when absent, escaped hostile version, `datetime` pattern on every page, no `<script>`, no `title`/`aria-label` additions, output without fields byte-identical to a rendering of the base templates (golden hashes recorded in T002). (US1.1-4, SC-003, SC-005)
- [ ] T010 [US1] Add `partials/freshness.html`; use it in `home.html` (card and index row) and `page.html` (both header variants); add CSS to `site.css`/`classic.css`; fix any test that hard-codes asset fingerprints. (FR-007, FR-008)
- [ ] T011 [P] [US1] CSS tests (`theme_cards_css_test.go` pattern): freshness rules use only palette tokens, no `display:none` in print, `flex-wrap` set, contrast of `--muted` on the card/surface pair holds in every palette and mode. (FR-008)
- [ ] T012 [US1] Real-Hugo test in `internal/render/hugo` (skipped without `SHEETMILL_HUGO`): build the same shuffled fixture in each theme with `--clock 2000-01-01T00:00:00Z` and `2031-06-01T00:00:00Z`; all files byte-identical; assert the `<time>` markup exists in built HTML. (US1.5, SC-002)

## Phase 4: Import and navi boundaries (US3)

- [ ] T013 [US3] Failing tests plus fixtures under `testdata/import/reference/freshness/` (valid, wrong-type, both with and without `cards`): archived under `source.original_metadata.cheatsheet`, `IMP112` positioned at the key, no active freshness, no inferred date, cards archive intact, finding order stable. (FR-010)
- [ ] T014 [US3] Implement the merged archival in `internal/importer/reference/reference.go` before `validate.Metadata`.
- [ ] T015 [P] [US3] Pair regression in `internal/render/navi/navi_test.go`: `.cheat` bytes, entry identities and diagnostics identical with/without the fields; companion keeps them verbatim; version is not a navi tag. Passes without navi code changes. (FR-011, SC-004)

## Phase 5: Contract, goldens, docs, demos

- [ ] T016 Add `testdata/reference-compatible/freshness.md` and generate its golden with `go test ./internal/markdown -run TestGolden -update`; inspect semantically and confirm `git diff --stat testdata` shows only the new files. (FR-014)
- [ ] T017 Update `schema/cheatsheet.schema.json` (`updated` pattern, `tool_version` bounds, `$comment` for calendar and invisible-character rules); add a schema test if the repo has a schema harness, otherwise a unit test that the properties exist.
- [ ] T018 Docs: `docs/format/gfm-plus.md` (field table, "Freshness" section, compatibility note), `docs/diagnostics.md` (`VAL004`, `IMP112` range), `docs/hugo.md` (shelf cards and header), `docs/imports.md`, `docs/format/navi-profile.md` (not exported, not a tag), `docs/cli.md` (`check` validates, `build hugo` shows, `build navi` unchanged). (FR-012)
- [ ] T019 Demo data: add `cheatsheet.updated` to all 15 `examples/sheets/*.md` and `tool_version` per [research.md](research.md); confirm `smill check` passes on the directory, `smill build hugo` works in three themes into a scratch directory, and `smill build navi` `.cheat` output is byte-identical to the base for all demo sheets. (FR-013, SC-006)

## Phase 6: Verification and handoff

- [ ] T020 Run `gofmt -l`, `go vet ./...`, `go test ./...`, `make check`, `make test-hugo` and `git diff --check`; report what ran and what did not (browser E2E/Node absent). (SC-001..006)
- [ ] T021 Record review rounds in `analysis.md`, push `codex/017-freshness`, hand over for the draft PR. No merge or deployment.

## Dependencies

T001 -> T002 -> T003 -> T004 -> T005 -> T006. US1: T007 -> T008 -> T009 -> T010 ->
T011/T012; T008 needs T004. US3: T013 -> T014; T015 is independent. T016 needs T004;
T017/T018 follow T004; T019 needs T005 and T010. T020/T021 last.
