# Tasks: Library navigation

**Input**: [spec.md](spec.md), [plan.md](plan.md), [research.md](research.md),
[data-model.md](data-model.md), [contract](contracts/navigation.md).
**State**: Implementation and local checks complete under the 2026-10-14 handoff.
T021 review and PR handoff are recorded below. CI readiness remains a separate
publication-state gate: do not mark the PR ready until its current checks pass.
Ordering fixtures are shared files; import fixtures preserve arbitrary upstream
values. No real library content was reordered.

`[P]` marks independent files that may be worked on concurrently once their stated
prerequisites pass. Story labels map to the specification. Write behavioral tests
first and observe the expected failure before implementing their corresponding rule.

## Phase 1: Scope and baseline

- [x] T001 Confirm explicit user implementation authorization and record its exact boundaries in `AGENTS.md`; retain an isolated `codex/` worktree and exclude deployment, PDF, navi execution and real-library reordering.
- [x] T002 Record the implementation base, existing parser/import golden checksums and baseline checks in `specs/014-library-navigation/quickstart.md`; inspect current renderer tests and the pinned `tests/e2e/` harness before extending them.

## Phase 2: Shared validation foundation

- [x] T003 Add failing table tests in `internal/validate/navigation_test.go` for absent navigation, empty mapping, optional fields, exact group preservation, whitespace-only group rejection, closed subkeys, invalid shapes/types, int64 orders 0/1,000,000, omitted versus zero, and out-of-range or float/bool/string orders.
- [x] T004 Implement shared `validate.Navigation` and derived `NavigationSettings` in `internal/validate/navigation.go`; return source-positioned VAL003 diagnostics, preserve input maps, and require canonical int64 order values without coercion.
- [x] T005 First add and observe failing source rejection/preservation integration cases in `internal/frontmatter/frontmatter_test.go`, then delegate `validate.Metadata` to the shared rules in `internal/validate/validate.go`; maintain existing diagnostics and unrelated metadata behavior.

**Checkpoint**: One shared rule implementation exists; no schema or metadata defaults
have been injected. All stories depend on T005.

## Phase 3: US1 — Place and order sheets (P1)

**Independent test**: A multi-sheet fixture produces identical home/sidebar group
membership and ordering for every tested input permutation.

- [x] T006 [US1] Add independently authored ordered, unordered, tied, overridden and uncategorized fixtures under `testdata/hugo/navigation/` and failing assertions in `internal/render/hugo/navigation_test.go`; cover zero versus omission, title/route ties and exact group identities.
- [x] T007 [US1] Implement derived settings, effective-group fallback and the member comparator in `internal/render/hugo/navigation.go` and `internal/render/hugo/hugo.go`; order explicit values first, numerically ascending, then lower-cased title, original UTF-8 byte title and route. Sort copied shelf members only and retain existing group ordering.
- [x] T008 [US1] Extend `internal/render/hugo/navigation_test.go` to compare all 24 permutations of a four-sheet case and seeded shuffles of the larger fixture byte-for-byte; assert sidebar/home agreement in both layouts and unchanged routes, related results, tag results, search and master page ordering.

**Checkpoint**: Placement works independently of categories, tags and source order.
Import safety and public source-contract documentation remain required before handoff.

## Phase 4: US2 — Diagnose source mistakes (P1)

**Independent test**: Invalid source and direct-IR inputs fail with consistent
VAL003 rules; valid navigation survives parsing without defaults.

- [x] T009 [P] [US2] Expand the source-positioned negative and preservation tests in `internal/frontmatter/frontmatter_test.go`, including YAML float `1.0`, nested unknown keys and no partial parse result on failure.
- [x] T010 [P] [US2] Add direct-IR acceptance/rejection and deterministic fallback-span tests in `internal/render/hugo/navigation_test.go`; cover malformed namespace values and canonical int64 versus float64 order, and ensure unrelated metadata validation is unchanged.
- [x] T011 [US2] Call the shared navigation validator at Hugo's direct-IR boundary in `internal/render/hugo/hugo.go`, after the existing nil/schema checks; return diagnostics before generating output.
- [x] T012 [US2] Add only `testdata/reference-compatible/navigation.md` and its new `.md.json` parser golden; inspect it semantically and verify every pre-existing parser/import golden remains byte-identical.
- [x] T013 [US2] Add the optional closed navigation object to `schema/cheatsheet.schema.json`; update `docs/format/gfm-plus.md`, `docs/hugo.md` and `docs/diagnostics.md` with field semantics, VAL003's wider meaning, ties, omitted order and source compatibility restriction. Explain YAML token typing versus JSON Schema integer semantics.

**Checkpoint**: Canonical source and direct callers share rules; source compatibility
is documented despite unchanged IR schema 1.

## Phase 5: US3 — Preserve import and export boundaries (P2)

**Independent test**: Incoming Reference navigation is inert provenance; canonical
navigation survives a navi companion without changing command bytes or identities.

- [x] T014 [P] [US3] Add failing import regression cases in `internal/importer/reference/reference_test.go` and new fixtures under `testdata/import/reference/` for valid and malformed upstream navigation; assert original-value archival, original key span and IMP102, with existing finding order and goldens unchanged.
- [x] T015 [US3] Archive and remove upstream Reference navigation before `validate.Metadata` in `internal/importer/reference/reference.go`; preserve existing visual-key archival timing and avoid inventing canonical placement.
- [x] T016 [P] [US3] Add regression tests in `internal/render/navi/navi_test.go` comparing otherwise identical documents with/without canonical navigation: companion metadata retains it exactly, `.cheat` bytes and exported-card identities remain identical. Do not invoke navi or exported commands.
- [x] T017 [US3] Update `docs/imports.md` for inert upstream navigation and its IMP102 finding; cross-link canonical navigation and companion preservation from `docs/format/gfm-plus.md` without claiming import activates placement.

**Checkpoint**: All three stories pass independently; no importer or navi exporter
has gained navigation behavior beyond the specified preservation boundary.

## Phase 6: Browser verification and final checks

- [x] T018 Extend the existing `tests/e2e/navigation.spec.mjs` suite using the GFM → Hugo → browser harness and the new fixture, preserving its tag/disclosure/palette regressions; verify sidebar/home agreement, current-group expansion, keyboard disclosure and sheet links in both layouts, root/subpath sites, desktop/mobile and JavaScript-disabled mode. Wait for loaded assets/fonts before geometry assertions; commands remain displayed content.
- [x] T019 Run the targeted source/import/Hugo/navi tests, full `go test ./...`, applicable race checks, `go vet ./...`, gofmt and `git diff --check`; confirm old golden bytes and generated route/card identities remain unchanged and run the existing browser/performance regression suites appropriate to this renderer change.
- [x] T020 Execute and record the actual local verification from `specs/014-library-navigation/quickstart.md`; reconcile this checklist and `specs/README.md` against evidence, leaving unrun checks pending and excluding content-library reordering.
- [x] T021 Obtain independent Sol review of the exact implementation revision, fix actionable findings with focused Conventional Commits, re-review, and attach the final evidence to the authorized PR; check CI before marking it ready. Do not merge or deploy.

## Dependencies and execution order

T001 → T002 → T003 → T004 → T005 is the shared foundation. US1 follows
T006 → T007 → T008. US2's T009/T010 may be written independently after T005,
then T011; T012/T013 finish source-contract coverage. US3's T014 precedes T015;
T016 is independent of import changes, while T017 follows T015 and T013.
T018 requires completed US1/US2; final T019–T021 require all stories and browser
coverage. T010 shares a file with US1: do not edit it concurrently with T006–T008.
T017 shares a documentation file with T013 and must follow it.

## Parallel examples

After shared validation, T009 (frontmatter tests) and T014 (import tests/fixtures)
can proceed in separate files. T016 (navi regression) can run alongside either.
After US1's renderer tests are finished, T010 can run alongside T009. Parallelism
is optional; it does not bypass dependencies or independent review.

## Delivery strategy

Implement and validate the shared foundation, then US1, US2 and US3 in focused
increments. US1 is the first demonstrable placement slice; all three stories and
final checks are required for the agreed first-slice handoff. Keep a future manifest,
group-order configuration, navigation title, stable IDs and deployment deferred.

## Review and PR handoff

[PR #33](https://github.com/wstein/sheetmill/pull/33) contains the implementation
and focused verification commits. Sol’s final review is clear at `e4c4572`;
[posted summary](https://github.com/wstein/sheetmill/pull/33#issuecomment-6165958344).
The final status-only commit does not change runtime/tests. CI has been inspected:
core passed at the reviewed head; web-e2e was running its final Studio checks.
The PR remains draft until all current-head checks pass; readiness evidence will
be posted on the PR. No merge or deployment.
