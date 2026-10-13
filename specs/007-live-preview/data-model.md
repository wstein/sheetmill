# Data model — 007

This introduces no database or canonical source/IR schema change. All session/run
models below are internal contracts. Stable diagnostic codes and source metadata
remain owned by the existing compiler.

## CorpusCase (implemented US1/US2)

- `id`: unique lowercase ASCII slug; cases are `smill`, `bdock`, `whr`.
  Default selection chooses only `smill`; explicit `all` chooses all three.
- `input`: repository-relative regular `.md` path; reject absolute paths, traversal
  and symlinks. MVP: `examples/sheets/smill.md`.
- `product`: display name and immutable verified source revision (full commit ID)
  with exactly 40 lowercase hexadecimal characters; unknown/missing identity is
  a fixture validation failure.
- `title`: nonempty exact expected page title.
- `search`: nonempty query, expected card ID, expected result label.
- `cards`: explicit expected stable anchors and representative semantic assertions.
- `copy`: selected card and explicit expected multiline text; preserve whitespace
  and trailing newline; do not trim or derive the oracle from generated output.
- `permalink`: required selected card ID and exact accessible link label.
- `attribution`: exact visible license/attribution expectations.
- `crossLinks`: array of `{label, sheet, card}` targets. Each target must identify
  a card in the selected corpus; missing targets fail before site generation
  or reader navigation.
  Empty for `smill`; BundleDock links to SheetMill semantic inspection, and
  WorkHarbor links to BundleDock volume options.
- Relationship: one case selects one source sheet; source frontmatter records its
  provenance. `corpus.json` is test configuration, never a new editable sheet format.

## JourneyRun (US1/US2)

- `runID`: internal identifier for owned temporary/evidence locations; excluded
  from site generation and canonical IR.
- `tools`: exact Go, Hugo, Node, Playwright and Chromium identities.
- `workspace`: newly created owned temporary directory, never a user repository.
- `stages`: ordered validate-tools → compile-test-cli → source-hash → check →
  inspect → export-two-projects → generate-two-sites → compare → browser →
  verify-source → cleanup.
- `processes`: known child handles with executable, fixed arguments, deadline and
  bounded output. Per-command logs ≤1 MiB; tool build deadline 120 s; Hugo deadline
  30 s; CLI invocation deadline 15 s after compilation. Setup allowance 300 s;
  individual reader test 120 s. Timeout/failure is explicit, not a silent retry.
- `violations`: attempted off-origin requests and forbidden process invocations;
  both lists must be empty for a successful journey.
- `evidence`: local stage logs, failure screenshot/trace; outside published trees.
- States: `preparing → generated → browsing → passed | failed → cleaned`.
  Failure in any earlier state still runs cleanup. No skipped prerequisite is a pass.

## GeneratedSite (US1)

- `files`: sorted normalized relative paths and bytes for every public regular file.
  Duplicate paths, symlinks or non-regular output entries are failures.
- `sourceHashes`: exact source bytes before and after; equality required.
- `clock`: fixed `2000-01-01T00:00:00Z` supplied to both Hugo invocations.
- `origin`: assigned `http://127.0.0.1:<ephemeral-port>`; never appears in generated
  bytes. Fixture serves a preloaded file map only; unknown routes return 404.
- Comparison: identical path sets and exact bytes; no normalization or exclusions.
  Cache and evidence locations are separate, so they cannot pollute output.

## SourceSelection (implemented US3/US5)

- `mode`: exactly `file` or `directory`; omitted CLI input means current directory.
- `root`: canonical explicitly selected parent/directory; never serialized to public
  output. Retain an `os.Root` handle and recheck entry type before every read.
- `entries`: eligible direct regular `.md` files; no recursion or symbolic links.
  A single-file selection grants access only to that file.
- Bounds: 100 selected files; 8 MiB per file; 32 MiB aggregate source bytes.
  Exceeding a bound is diagnosed and no selection is silently truncated.

## SheetRevision (implemented US3/US4)

- `sheetID`: existing route identity; duplicate routes mark every origin invalid.
- `sequence`: monotonically increasing session revision; local, never canonical IR.
- `sourceHash`: SHA-256 of saved bytes; reads must represent a stable snapshot.
- `diagnostics`: existing source-positioned compiler/renderer diagnostics.
- `currentIR`: validated schema-1 IR or absent on invalid source.
- `lastValidIR`: optional previous validated IR used only with an explicit stale label.
- `status`: `pending | current | invalid | removed | failed`.
- Transition: observed change → pending → current on successful latest render;
  pending → invalid/failed retains last-valid visibly; deletion → removed.
  Removed sheets are not served as current pages. Older completions are discarded.
  Valid sheets remain independently usable when another selected sheet is invalid.

## PreviewSession / BrowserState (implemented US3–US5)

- `sessionID`: random local identity; never canonical output or source metadata.
- `listener`: IPv4 loopback only; one selected port, default dynamically assigned.
- `authorization`: cryptographic one-use bootstrap secret and session cookie;
  private state, excluded from logs/traces and source output. No project-write capability.
- `stateVersion`: monotonic across pending/publication transitions, independent
  of source sequence, so delayed state responses cannot regress the UI.
- `servedPublication`: trusted HTTP HTML identity compared with private state on
  initial fetch and reconnect; canonical generated bytes remain unchanged.
- `currentSequence`: latest observed revision; `publishedSequence` can differ only
  when explicitly pending/invalid/failed.
- `snapshot`: immutable generated files plus per-sheet freshness; one in-flight
  publisher and at most one queued newest revision.
- `connection`: browser `connected | disconnected | reconnecting`; disconnect
  immediately makes freshness unknown and visible, never silently current.
- Bounds: 64 MiB generated bytes total, including retained last-valid state; 8
  browser streams; newest pending revision replaces older queued work.
- Shutdown: stop watchers/streams, cancel known publishing child, bounded 5 s
  graceful cleanup, force only that owned child if necessary; close listener/root
  handles and remove owned temporary state. Never broad-kill processes or delete sources.
