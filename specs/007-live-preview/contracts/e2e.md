# E2E harness contract — MVP and additional inputs

Status: implemented for the one-sheet MVP and US2 additional-input corpus.
Read-only server journeys are implemented in `studio.spec.mjs` under the
separate `studio.config.mjs`, with authenticated tracing disabled. Maps to US1/US2, FR-001/004/009/010 and SC-001/004/005.

## Invocation and prerequisites

From `tests/e2e`, `npm test -- --project=desktop` runs the complete first journey;
`npm test` additionally covers mobile and JavaScript-disabled contexts. The package
is private and has no publish script. Exact Node 24.21.0 and Playwright 1.64.0 are
checked before running; browser comes from the matching installed Playwright cache.
`SHEETMILL_HUGO` must be an absolute path to the installed Hugo 0.167.0 executable.
No download/install is performed by tests; missing prerequisites fail clearly.

`corpus.json` contains three independently authored cases: `smill`, `bdock` and
`whr`, mapped to their files under `examples/sheets/`. The default selection is
exactly `smill`; `SHEETMILL_E2E_CORPUS=all npm test` selects all three together.
Other selection values fail. Every selected case uses the same shared reader
assertions, with explicit per-case copy/permalink oracles and cross-card targets.
No sibling runtime dependency is introduced. Each case follows
[data-model.md](../data-model.md).

## Process and filesystem boundary

Allowed tool invocations are fixed by fixture code:

1. Go tool version and temporary test compilation of `./cmd/smill` (`-trimpath`, no
   release flags, output in an owned tools directory; never committed/uploaded).
2. The resulting executable: `version`, `check <selected-source>`, `inspect
   <selected-source>` and `build hugo <selected-source> --out <new-project>`.
3. Explicit Hugo executable: `version`, then `--source <owned-project> --destination
   <owned-public> --clock 2000-01-01T00:00:00Z --noBuildLock --quiet`.

No shell, `bdock`, `whr`, navi, Typst, Docker, source snippet, arbitrary executable
or user-provided argument string is invoked. Repeated site generation uses distinct
new directories. Temporary source paths resolve symlinks in system temp roots first,
so existing strict source-path checks still apply on macOS. Original committed
sheets are never modified. Cache/environment paths are owned and are not serialized.

The static fixture server binds only `127.0.0.1:0`, serves a generated-file map with
known MIME types and directory-index routes, disables listings and returns 404 for
unknown/malformed paths. No request-triggered filesystem traversal. Never reuse
`127.0.0.1:1313`, an already running server or an installed `smill`.

## Browser journey and oracles

Shared scripted flow in both desktop and mobile projects, using the same oracle:

1. Load generated home; assert title, the selected reference count and the case's
   exact reference link.
2. Enter the corpus query; assert the expected result and accessible result label.
3. Press Tab; assert focused result; press Enter; assert the exact card fragment,
   heading and expected command text.
4. Seed the clipboard with a sentinel. Click that card's real copy button, read the
   actual clipboard under permissions scoped to this origin, and compare exact
   expected multiline bytes. Do not stub clipboard APIs or accept only a toast.
5. Switch light/dark modes; inspect the actual applied state and reload to prove
   preference persistence; navigate a stable card link and verify its destination.
6. Verify attribution, no browser errors, unchanged source, an empty forbidden-
   request list and `documentElement.scrollWidth <= documentElement.clientWidth`.
   Do not execute copied text.
7. Follow every declared cross-sheet/card link and verify the destination URL,
   target heading and card. The no-JavaScript context also checks these links.

The complete shared flow above MUST pass at both desktop 1280×800 and mobile
390×844: search, keyboard result activation, exact clipboard bytes, appearance
change/reload persistence, card links and page overflow are not desktop-only
checks. A distinct `javaScriptEnabled:false` context verifies the real sheet/card
content and ordinary links without expecting search/copy/live refresh enhancements.
Only that no-JavaScript context receives enhancement exemptions; the scripted
mobile project receives none. No pixel-baseline test is required for the MVP.

Before any navigation, context routing permits only the exact fixture origin;
service workers are blocked. Every blocked off-origin request is also recorded
and fails the test. No general external-origin exception or silent network stub.
Use semantic locators and auto-retrying assertions; no fixed-duration sleeps.

## Determinism, evidence and cleanup

Before browser interaction, compare two complete sorted public file sets and all
corresponding bytes. Compare source hashes before and after the full run. Evidence
is outside both public trees: ignored `.local/e2e/` by default, configurable only
as a local output path. Logs ≤1 MiB per child; include stage/executable/outcome,
redacting credentials. Zero retries; traces retained on failure and screenshots
only on failure. Capture setup logs even if the browser was never launched.

Cleanup always closes owned contexts/listener, stops and awaits only known children
with bounded escalation, and removes only the fixture's owned temporary trees.
Keep evidence separately. Interruptions and failed setup are covered by harness
self-tests. One worker initially; no external reporter, hosted trace viewer or
artifact upload. CI stores no generated site, executable or evidence as a published
artifact. Required E2E CI runs all contexts and cannot treat a skip as success.
