# Validation guide and evidence — 014

Implementation authorized 2026-10-14 in AGENTS.md. Worktree:
`/private/tmp/sheetmill-navigation-impl-20261014`. Original production implementation
was inherited as commits `fb14f79` and `e08e72f`; verification additions follow.
The author’s test-first history was not independently reconstructed.
Baseline: main `de3e08e55c5e42a1eedc4ea24e749219ae35f1f5`.
[Checksums](baseline-checksums.json) record every pre-existing parser/import fixture;
all match byte-for-byte. The following remains a reproducible acceptance guide.

1. Add independently authored multi-sheet fixtures covering group overrides,
   category fallback, uncategorized, omitted/zero/cap orders, duplicate orders,
   case/title/route ties and related relevance ties. Keep source filenames unchanged.
2. Run source negative tables for mappings, group content, all order types/ranges
   and unknown keys. Confirm VAL003 key spans. Repeat with direct Hugo IR and
   confirm 1:1 fallback plus no partial project.
3. Add `testdata/reference-compatible/navigation.md` and its metadata-only golden.
   Explicitly generate only the new subtest, then verify all existing goldens are
   byte-identical to the pre-implementation revision:

   ```sh
   go test ./internal/markdown -run '^TestGolden/navigation\.md$' -update
   go test ./internal/markdown -run TestGolden
   ```

4. Build the multi-sheet fixture with both layouts, root/subpath and configured
   canonical origin. Compare group/member sequences between home and sidebar.
   Run all 24 permutations of a four-sheet subfixture and seeded shuffles of the
   full fixture, asserting byte-identical exported projects.
5. Convert Reference inputs containing valid and malformed navigation. Assert
   root navigation absent, original value preserved in provenance, IMP102 at its
   original position, and unchanged unrelated import finding order/goldens.
6. Compare navi text exports with/without authored navigation: `.cheat` bytes and
   exported-card identities match; companion metadata preserves navigation.
   Use the in-process exporter tests; never invoke the navi executable or snippets.
7. Run the existing actual-Hugo browser harness with desktop/mobile/no-JavaScript,
   both layouts and both mounts. Wait for page/styles/fonts before layout checks.
   Verify disclosure/current-page behavior still works after ordering.
8. Run gofmt, `go test ./...`, `go test -race ./...`, `go vet ./...`, applicable
   browser suites and `git diff --check`. Record exact revision, actual results
   and limitations, then obtain independent review. Do not merge or deploy.

## Local results (2026-10-14)

- `go test ./...`, `go test -race ./...`, `go vet ./...`: pass.
- `SHEETMILL_HUGO=/opt/homebrew/bin/hugo go test ./internal/render/hugo`: pass,
  including real Hugo integration in both layouts.
- Complete reader/corpus Playwright suite: 45/45 pass. This includes 12 authored
  placement journeys across two layouts, two mounts and three browser profiles,
  plus existing tags, palettes, assets, copy, inert-input and product-sheet journeys.
- Node harness/performance-unit/search checks: 23/23 pass.
- Ordering checks cover all 24 four-sheet permutations and 50 seeded full-fixture
  shuffles, Unicode/title/route ties and exact group identity. Source checks assert
  nested 4:3 positions and no partial parse; imports compare complete values.
- Only new fixtures and one new parser golden were added; prior fixture/golden
  bytes are unchanged. Metadata companions preserve valid navigation without
  changing `.cheat` bytes or card identities; navi was not executed.
- gofmt and `git diff --check`: pass. No extension hooks configured.

The first performance attempt stopped at the harness’s clean-tree guard after
startup sampling because verification edits were not committed. No acceptance
result is claimed for that attempt; the committed-input rerun is recorded below.
