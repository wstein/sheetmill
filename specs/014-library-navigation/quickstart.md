# Future validation guide — 014

These checks describe implementation acceptance; they have not been run for a
014 implementation. Start only after implementation scope is explicitly recorded
in AGENTS.md. Use an isolated worktree and existing pinned Go/Hugo/Node/Playwright.

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
