# Quickstart validation — planned 007

This is a run guide for the future implementation. The E2E package, authored
`smill.md`, CI job and `serve` command do not exist merely because this plan does.
Current core/Hugo tests remain runnable; proposed commands below become valid when
the corresponding tasks are complete. No PDF, navi or deployment steps appear.

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

After implementation, from the worktree root:

```sh
export SHEETMILL_HUGO=/absolute/path/to/hugo
cd tests/e2e
node --version
npm ci
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

## Later stages, separately selected

After the one-sheet gate, add pinned/verified `bdock` and `whr` corpus cases and run
`SHEETMILL_E2E_CORPUS=all npm test` from `tests/e2e`. Those product binaries must not
be installed/invoked by the harness. Verify representative cross-sheet links.

After the server-stage tasks, from a worktree root:

```sh
go run ./cmd/smill serve examples/sheets/smill.md --open --hugo /absolute/path/to/hugo
```

Save changes from an external editor. Observe new content, then invalid-source
diagnostics, repair, reconnect and latest-save behavior. This proposed command is
not part of the first slice. The browser remains read-only. Run the later studio
journeys and Go security/lifecycle tests described in [serve.md](contracts/serve.md).
Measure SC-002 on the documented workstation and SC-006 with five authors; do not
claim those outcomes from the static first-slice journey.
