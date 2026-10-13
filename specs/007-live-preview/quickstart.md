# Quickstart validation — 007 publishing and read-only Studio

The one-sheet GFM → CLI → Hugo → browser MVP, US2 corpus and read-only
`smill serve` increments are implemented. No PDF, navi, browser editor or
playground is implemented. See [publishing evidence](../../docs/e2e-verification.md)
and [Studio checks/gaps](../../docs/studio-verification.md).

## Prerequisites

- Go 1.27.2 (repository pin).
- Installed Hugo 0.167.0; provide its absolute executable path via `SHEETMILL_HUGO`.
- Node 24.21.0 LTS for `tests/e2e`, with npm and the committed dependency lockfile.
- Matching Playwright 1.64.0 Chromium installed explicitly before offline tests.
  Linux needs its documented browser system dependencies; provision those during
  setup, not by the test process. No sibling CLI, container or account is required.

All operations use your own feature worktree. On macOS resolve `/tmp` to
`/private/tmp` when selecting source/output locations, preserving symlink rejection.

## First slice: one sheet, one complete publishing journey

From the worktree root:

```sh
export SHEETMILL_HUGO=/absolute/path/to/hugo
cd tests/e2e
node --version
npm ci --ignore-scripts
npx --no-install playwright install chromium
npm test -- --project=desktop
npm test
```

`node --version` must report v24.21.0. Dependency/browser installation above is
explicit setup; subsequent `npm test` runs must not access the network for downloads.
The desktop command proves the complete one-sheet reader journey at desktop size.
The full command repeats every scripted reader assertion at mobile size and adds
a separate no-JavaScript context. Corpus selection defaults to `smill` only.
Do not replace the fixture with the user-facing server already running on port 1313.

Expected outcomes: check/inspect/export run through the actual CLI; two complete
Hugo outputs match byte-for-byte; source bytes are unchanged; search and Tab/Enter
reach the expected card; actual copied bytes match the oracle; appearance survives
reload; both scripted viewports pass search, keyboard activation, exact copy, appearance
persistence and no page overflow; basic no-JavaScript navigation works; no
external requests or example executions occur. The runner cleans owned processes
and temporary directories. See [E2E contract](contracts/e2e.md) for exact assertions.

On failure, inspect ignored local `.local/e2e/` stage logs and retained failure
screenshots/traces. Open traces only in the local viewer:

```sh
npx --no-install playwright show-trace /absolute/local/path/to/trace.zip
```

Do not upload site output, traces, test executables or source contents. A failure
before browser startup should still have the relevant tool-stage output.

## Existing regression checks

From the worktree root, these commands already exist:

```sh
go test ./...
go test -race ./...
go vet ./...
SHEETMILL_HUGO=/absolute/path/to/hugo go test ./internal/render/hugo -run TestHugoIntegration -count=1
```

A skipped optional Go integration test is not a substitute for the required E2E job.

## Additional inputs: implemented US2

Run `SHEETMILL_E2E_CORPUS=all npm test` from `tests/e2e` for pinned/verified
`smill`, `bdock` and `whr` references in one library. The default remains `smill`
only. Those product binaries are not installed/invoked by the harness. The shared
journeys verify cross-sheet/card links, search, exact copy and attribution.
See [corpus identities and evidence](../../docs/e2e-corpus.md).

## Read-only saved-file Studio: implemented

From a worktree root, with installed pinned Hugo:

```sh
go run ./cmd/smill serve examples/sheets/smill.md --open --hugo /absolute/path/to/hugo
```

Save changes from an external editor. Observe new content, then invalid-source
diagnostics, repair, reconnect and latest-save behavior. This command belongs to the separately authorized read-only server increment. The browser remains read-only. Run `npm run test:studio` from `tests/e2e` for authenticated desktop/mobile
journeys; traces are disabled. The same installed-tool prerequisites apply.
Use `SHEETMILL_HUGO=/absolute/path/to/hugo go test -race ./internal/studio` for
Go security/lifecycle/integration tests described in [serve.md](contracts/serve.md).
SC-002/SC-003 have opt-in automated reference-workload checks; see
`npm run test:performance` and `docs/studio-verification.md` for measured evidence.
SC-006 five-author usability remains pending, so T049 stays unchecked overall.
