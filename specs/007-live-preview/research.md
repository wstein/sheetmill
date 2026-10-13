# Research — 007 local live preview and its E2E first slice

Date: 2026-10-13. Repository inspected after rebase onto current `origin/main`.
Research was read-only; no tools installed, dependencies changed or test harness run.
A separately dispatched research agent verified the browser-tool decisions using
primary sources. All decisions below are proposed design, not measured results.

## 1. Prove the existing publishing path before the server

**Decision**: First implement one `smill` sheet and a complete actual CLI → Hugo →
browser journey. Compile the real `cmd/smill` entry point to a temporary test
executable; invoke it directly with `shell:false` and fixed argument arrays.

**Rationale**: Existing CLI tests call `cli.Run`; existing Hugo integration calls
`hugo.Render` and actual Hugo. Both are useful but neither alone tests the executable,
export IO and reader workflow together. A temporary test executable avoids an
installed binary or long-lived `go run` descendant and is not a release build.

**Alternatives considered**: Component-only tests, browser HTML mocks, an installed
`smill`, or implementing `serve` first all miss the selected slice or widen it.
[Node subprocess API](https://nodejs.org/api/child_process.html) documents direct
executable invocation and process termination limitations. Use no intermediary shell.

## 2. Pin the development browser stack

**Decision**: Isolated `tests/e2e` package; Node 24.21.0 LTS; `@playwright/test`
exactly 1.64.0 and a committed npm lockfile. Use its bundled Chromium revision 1248
(version 156.0.8078.4), not the user's installed Chrome. npm installation and browser
installation occur explicitly before tests. No production JavaScript compiler/UI.

**Rationale**: Versioned browser behavior and tool reproducibility matter for
clipboard permissions, layout and browser assertions. Registry metadata declares
Node ≥20; Node 24 is a supported LTS line. The local default Node 26.11.1 observed
in this session is not silently adopted as the pinned CI toolchain.

**Alternatives considered**: Floating package versions, system Chrome, multiple
browser engines initially, Go browser bindings or an entire frontend framework.
Playwright offers the needed runner, contexts, request routing and traces without
expanding the Go runtime dependency graph.

Sources: [exact package metadata](https://registry.npmjs.org/@playwright/test/1.64.0),
[versioned browser manifest](https://raw.githubusercontent.com/microsoft/playwright/v1.64.0/packages/playwright-core/browsers.json),
[Node 24.21.0 release](https://nodejs.org/en/blog/release/v24.21.0),
[Node support schedule](https://raw.githubusercontent.com/nodejs/Release/main/schedule.json).
CI action references or installer checksums must be resolved and verified before
implementation commits; no floating action tag or tool download during tests.

## 3. Own the whole temporary environment

**Decision**: One Playwright worker, no retries. Fixture owns source copies, two
export roots, caches, a generated-file-only in-process HTTP server on loopback port
0 and all child processes. No `reuseExistingServer` and no fixed port 1313.

**Rationale**: This cannot accidentally pass against the user's running preview,
share an export directory or execute stale output. Single-worker operation reduces
clipboard cross-test interference. Clean up contexts/listeners/children on success,
setup failure, timeout and interruption; delete only owned temporary trees.

**Alternatives considered**: Reuse Hugo's development server, a Python server,
fixed ports, shell scripts and broad process-kill commands. None is necessary for
the static first slice. [Playwright server guidance](https://playwright.dev/docs/test-webserver)
explains server reuse; this harness instead owns a dynamic-address fixture directly.

## 4. Verify outcomes instead of UI acknowledgements

**Decision**: Store explicit expected titles, anchors, source attribution and
multiline snippets in the corpus oracle. Assert keyboard focus and destination,
actual clipboard text, appearance persistence and overflow. Separate contexts
cover mobile and JavaScript-disabled navigation; use semantic locators and
condition-based waits.

**Rationale**: A “Copied” label, screenshot or parsed source alone does not prove
that a reader received the right command. Grant clipboard permissions only for
the fixture origin; seed a sentinel, click the actual control, then read back.
No clipboard-function mocking or altering site source to make the test pass.

**Alternatives considered**: Broad image snapshots, fixed sleeps, click-only tests,
or deriving all expected output from the renderer under test.
Sources: [assertions](https://playwright.dev/docs/test-assertions),
[context options](https://playwright.dev/docs/test-use-options),
[permissions](https://playwright.dev/docs/api/class-browsercontext#browser-context-grant-permissions).

## 5. Offline boundary and deterministic comparison

**Decision**: Install a context-wide request guard before navigation; allow only
the exact fixture origin, block service workers and record every attempted
non-local request as a test failure. Compare every corresponding public file byte
and sorted relative file list from two independent builds with fixed clock
`2000-01-01T00:00:00Z`. Hash source bytes before and after the journey.

**Rationale**: Blocking alone can conceal unintended requests; assertion catches
them. Test outputs/caches/evidence are outside public trees, not exclusions that
weaken comparison. Source snippets are never promoted into commands to execute.

**Alternatives considered**: Public playground requests, online themes, normalized
HTML-only comparisons or evaluating all example commands.
Sources: [request interception and service workers](https://playwright.dev/docs/api/class-browsercontext#browser-context-route),
[Hugo command/clock](https://gohugo.io/commands/hugo/).

## 6. Local failure evidence without publication

**Decision**: Zero retries, trace `retain-on-failure`, screenshots `only-on-failure`,
line reporter and bounded stage logs; evidence under ignored `.local/e2e/` by
default. CI may retain evidence inside its job workspace for investigation but
adds no upload-artifact step or hosted trace link. Logs and tokens are redacted.

**Rationale**: `on-first-retry` produces no trace with zero retries. A setup failure
needs subprocess logs even when no browser was reached. Traces can contain source
content and must not become implicit publication. Later authenticated Studio
contexts disable tracing entirely: network traces can include session cookies even
after bootstrap. Use failure screenshots and sanitized stage logs there; session
credentials never enter reports.

**Alternatives considered**: External reporters, unconditional videos/screenshots,
automatic artifact uploads or opening traces in a hosted viewer.
Sources: [recording options](https://playwright.dev/docs/test-use-options),
[local trace viewer](https://playwright.dev/docs/trace-viewer).

## 7. Later server design remains separate

**Decision**: Native Go standard-library HTTP, read-only selected-file watching,
content-hash polling and server-sent revision events. Rebuild with the existing
compiler/renderer and the explicitly selected pinned Hugo tool in owned snapshots.
No filesystem writes from the browser; no arbitrary command endpoint.

**Rationale**: Polling avoids an unneeded platform watcher dependency and handles
atomic saves; one queued latest revision prevents obsolete output. SSE fits
one-direction notifications. Origin/host checks must protect GET reads too;
standard CSRF middleware alone does not authorize private read access.

**Alternatives considered**: WebSockets, a new JS compiler, rich-text editing,
Hugo's own unsupervised watcher, or a remote compilation service. These widen the
milestone or bypass its selection/security model.
Sources: [Go HTTP lifecycle](https://pkg.go.dev/net/http#Server.Shutdown),
[os.Root containment](https://pkg.go.dev/os#Root),
[CrossOriginProtection limitations](https://pkg.go.dev/net/http#CrossOriginProtection).

No unresolved design questions remain for planning. The sibling corpus identities,
lockfile integrity and installer/action digests are explicit later verification
work, not assumptions that any currently installed product is authoritative.
