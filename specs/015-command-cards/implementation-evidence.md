# Implementation evidence — 015

Implemented in isolated branch `codex/015-command-cards` under the user’s
2026-10-14 T001–T011 handoff. Baseline main `9f132df`; fixture hashes are in
[baseline-checksums.json](baseline-checksums.json). No navi, shell/source snippet,
provider, PDF or deployment was run. Host shells ran development tools only.

## Test-first and identity evidence

- Core subagent observed `.command` rejection and invalid shared references
  accepted before implementing parser/metadata/document rules. Source diagnostics,
  exact nested spans, direct IR, nonmutation and legacy boundaries are tested.
- Navi subagent observed new cards skipped, namespace conflicts ignored and mixed
  references rejected before integration. Tests pin complete output and two entry
  hashes: `0x362d24b0c6381291` and `0x1d9f85deaea8b6e6`.
- The pinned legacy test independently passed against the **pre-change exporter**
  in detached baseline `9f132df`, using only the new regression test source:
  `go test ./internal/render/navi -run '^TestCommandMigrationPinnedOutputAndCompanions/legacy$' -count=1`.
  This compares actual old output with fixed bytes and hashes; it does not run navi.
- All four profiles and all five syntax variants compare complete fixed `.cheat`
  bytes (only the selected profile header varies across profiles) and pinned entry
  hashes. NAV008, existing hash vectors, contexts, order, strict atomicity and raw
  companion metadata are covered. A CLI regression parses actual new Markdown.
- Import subagent observed invalid upstream shared metadata failing core checks,
  valid metadata remaining active and old marker diagnostics before inert archival.
  IMP111 tests compare complete values and original marker/key positions. Only a
  new import fixture and its canonical/report golden were generated and inspected.
- All **28 pre-existing testdata files** still match their baseline SHA-256 bytes.
  Existing golden files were not regenerated. IR schema and navi output version stay 1.

## Actual local verification

At runtime/test commit `d8cd3b0` (later changes only reconcile documentation):

- `go test ./...`: pass, including manual export and no-execution renderer lint.
- `go test -race ./...`: pass.
- `go vet ./...`: pass.
- `SHEETMILL_HUGO=/opt/homebrew/bin/hugo go test ./internal/render/hugo`: pass.
- Existing complete reader/corpus Playwright suite: **45/45 pass**, covering
  desktop/mobile/no-JavaScript, both layouts, both mounts, inert examples and the
  three product sheets. No new visual behavior is introduced.
- `python3 -m unittest discover -s assets -p 'test_*.py'`: 3/3 pass.
- gofmt, baseline checksum assertions and `git diff --check`: pass.

Tool versions remain the repository pins. Performance acceptance and the
five-author usability study are not claimed by this increment. No extension hooks
are configured. CI and exact-final-head review are recorded on the implementation PR.

## Independent review

A separately assigned Sol reviewer inspected core/shared validation, schema,
source/direct-IR boundaries, aliases/conflicts, inert imports, companions and
identity/output preservation, and independently ran targeted/full Go tests.
No blocking correctness finding was found. Both recommendations were fixed and
re-reviewed: NAV104 now guides shared cards to the new namespace while preserving
legacy wording; every profile now has fixed-byte/hash parity across all variants.
Code re-review is clear at `d8cd3b0`; final-head confirmation is posted on the PR.

`checklists/requirements.md` is preserved read-only as the initial requirements
quality review from docs PR #34 / baseline `9f132df`. Its unimplemented/runtime-
pending note describes that earlier documentation revision. This evidence and
`tasks.md` record the subsequent implementation; the old checklist is not used
as a claim about current runtime verification.
