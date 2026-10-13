# 007 MVP verification — 2026-10-13

Scope: T001–T018 and MVP portions of T047/T048/T050. All edits are in the author's
isolated worktree on `codex/007-mvp-e2e`. One real-product reference,
`examples/sheets/smill.md`, is independently authored against CLI revision
`0f87dfb7b2fc5a3269e76f6c6f579cb863e430a3` and read-only help checks. Its eight
cards document implemented behavior, not a future `serve` command. Temporary
adversarial variants are not additional product corpus cases.

## Actual local environment and commands

macOS arm64; Go 1.27.2; Node 24.21.0; npm 11.19.0; Playwright 1.64.0;
Chromium revision 1248 / 156.0.8078.4; installed Hugo 0.167.0 extended (Homebrew).
Node was checksum-verified; dependencies/browser were explicitly installed in
advance. Browser cache and pinned Node were kept in temporary storage.

Executed from the isolated worktree/test package with pinned Node on PATH,
`PLAYWRIGHT_BROWSERS_PATH` pointing to that installed cache,
`SHEETMILL_HUGO=/opt/homebrew/bin/hugo`, and
`GOCACHE=/private/tmp/sheetmill-go-cache`:

| Check | Actual result |
| --- | --- |
| `node --test tests/e2e/harness.test.mjs` | 14 passed after review corrections |
| Playwright desktop project | Complete reader journey passed |
| `npm test` | 6 passed; full desktop/mobile and separate no-JavaScript contexts, plus inert variants |
| Repeated full suite, fresh owned workspaces | 6 passed again; every pair of public trees identical |
| `go test -race ./...` with explicit pinned Hugo | Passed, including real Hugo integration |
| `go vet ./...`, `go mod verify` | Passed; all modules verified |
| Empty `gofmt -l cmd internal assets`, `git diff --check` | Passed |

Every scripted desktop and mobile reader journey verifies keyboard search/result
activation, exact card/command content, actual clipboard text including trailing
newline, light/dark applied colors and reload persistence, stable links,
attribution, zero page overflow and no browser errors. No clipboard stub or
“Copied” acknowledgement substitutes for readback. No-JavaScript checks retain
ordinary content/navigation and a disabled search input. Inert variants verify
escaped HTML, literal shortcodes, unfetched external images and an uncreated shell
marker. All browser contexts block service workers and fail on off-origin requests.

Test-tool stage records verify local/offline Go settings, including the very first
prerequisite invocation. No test automatically installs a tool or browser. The
required E2E workflow provisions prerequisites in explicit setup steps and then
runs all contexts without permissive missing-tool skips. Hosted Linux results are
reported separately in PR checks; local results are not presented as hosted CI.
No branch-protection settings are changed by adding the `web-e2e` workflow.

## Determinism and ownership

Successful baseline snapshots contained **14 public files**. Sorted UTF-8 relative
filename + NUL + bytes + NUL yielded SHA-256:
`580610e9d9b6a409cfac4ed8e431b1749d1d60098de2cb888d8200d6c9d1a5a4`.
This digest matched across independently generated baseline sites. Source SHA-256:
`4d3228787a9570ac1cdb85207e458f590c70255c57b3680006d4e9ff99310f27`.
Original and copied source hashes remained unchanged throughout each journey.

The fixture compiles the real CLI into an owned temporary tools directory and
serves only a preloaded generated-file map on an ephemeral IPv4 loopback port;
it never uses the user's running preview. Every public filename and byte is
compared without normalization/exclusions. Source examples never select executable
arguments. Temporary executables, source copies, listeners and children are
cleaned up; bounded local evidence lives separately in ignored `.local/e2e/`.
No artifact upload, external reporter or publication is configured.

## Test-first and review evidence

The helper suite initially failed with missing helper modules; the complete reader
test failed with a missing fixture before implementation. Later, Sol independently
reproduced two P2 defects: competing per-owner interruption exit handlers, and Go
prerequisite execution before offline settings. Focused regressions were observed
failing, then passing after correction. A coordinated signal handler now awaits
all owners through temporary-tree removal. Offline/local Go settings are constructed
before all tool calls; unit and actual journey assertions verify that boundary.

Sol re-reviewed exact source/CI revision
`2c4f8325ac8034dc2deb8e973133c32a07fc9c7e`: both findings closed, with no further
source/CI findings. Independent fresh-snapshot checks passed 14 helpers and all
6 browser journeys. Final documentation reconciliation receives a separate
exact-HEAD review before the implementation handoff is marked ready.
The first review independently passed the then-current 12 helpers and 6 browser
journeys. Those are not claims that the later Studio stories are implemented.

## Remaining scope

The original MVP added no `smill serve`, browser editor, sibling-product corpus,
playground, PDF, navi, deployment, release or distribution build. The separately
authorized US2 corpus increment below adds only sibling sheets and harness
coverage; T019–T041 and T049 remain unchecked. Live-update latency, five-author usability, other browser engines and
Windows process lifecycle are not measured or claimed. Exact site bytes across
different operating systems are not asserted; repeated builds use identical tools
and configuration on each host. Independent review and hosted CI results remain
separate evidence from author-local checks.


## US2: additional inputs (2026-10-13)

T042–T046 add independently authored BundleDock and WorkHarbor references at
verified immutable source revisions. WorkHarbor follows the user-selected
`workharbor-nxt` bootstrap: its sheet makes no implemented `whr` CLI claim.
After that source choice, helper/default tests and both all-corpus runs were
repeated successfully. The default `smill` suite passed all 6
checks; the explicit `SHEETMILL_E2E_CORPUS=all` suite passed all 12 checks twice
from fresh owned roots. All 15 helpers passed, along with Go race tests, vet,
module verification, formatting and diff checks. Both sibling executables were
absent from the test PATH. No source example was executed.

The same reader assertion function covers all products at desktop/mobile sizes
and with JavaScript disabled. Cross-sheet/card targets, attribution and exact
clipboard bytes are tested; every generated tree and selected source hash is
compared. See [corpus inventories, hashes and evidence](e2e-corpus.md). The existing
[deployed demo](deployed-demo-verification.md) was verified read-only before this
increment. Existing Pages workflow behavior was not modified.

The new selection/target helper test was observed failing against the MVP-only
loader before implementation, then passing with explicit selection and cross-card
validation. Additional permalink validation rejects a missing oracle before a
browser journey begins. Source/CI and final documentation receive independent
Sol review; exact review results are recorded in the corpus evidence document.


## Read-only Studio increment

T019–T041 were subsequently authorized and implemented. See
[Studio verification](studio-verification.md) for actual Go/browser results,
review fixes and updated sheet hashes, and [security boundary](studio-security.md)
for tested request/source/process constraints. The source-relative diagnostic and
freshness journeys run against the real `smill serve`, compiler and installed
Hugo. The original static MVP/US2 evidence above remains historical; Studio did
not add PDF, navi, browser editing or deployment. T049 remains unmeasured.
