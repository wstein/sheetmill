# Reference web E2E

Development-only package: Node **24.21.0**, Playwright **1.64.0**, bundled Chromium
revision **1248** (156.0.8078.4), Go **1.27.2** and installed Hugo **0.167.0**.
No sibling CLI, container or account is required. No production JavaScript compiler.

Explicit setup, from this directory:

```sh
npm ci --ignore-scripts
npx --no-install playwright install chromium
export SHEETMILL_HUGO=/absolute/path/to/hugo
npm run test:harness
npm test -- --project=desktop
npm test
```

Use the pinned Node on PATH. Linux browser libraries must be provisioned explicitly
before testing (CI uses Playwright's documented `install --with-deps chromium`).
After setup, test execution downloads nothing; it fails on missing prerequisites.
`PLAYWRIGHT_BROWSERS_PATH` may select a preinstalled matching browser cache.

The harness owns a temporary real CLI executable, source copies, two independent
Hugo projects and a loopback server on an ephemeral port. It never uses an installed
`smill` or the user's preview on port 1313. It invokes only fixed Go build/version,
smill version/check/inspect/build-hugo and pinned Hugo version/build arguments via
direct subprocess APIs. Commands displayed in sheets are content, never executed.

The default selection remains the single `smill` case. Set
`SHEETMILL_E2E_CORPUS=all npm test` to generate `smill`, `bdock` and `whr`
together and exercise every sheet through the same journey. Unknown selections
fail explicitly. Cross-sheet links are checked against selected card identities. Desktop and mobile run
the same full reader journey: keyboard search/navigation, actual clipboard contents,
appearance persistence, card links and overflow. A separate no-JavaScript context
checks ordinary navigation/content. Temporary adversarial variants of the same
sheet prove source HTML/shortcodes/snippets stay inert. Off-origin browser requests
are blocked **and fail the test**. Source hashes and complete built file bytes are
compared. No broad visual snapshot baseline is used.

Evidence is local and ignored under `.local/e2e/`: bounded tool-stage logs plus
failure screenshots/traces. Setup failures have logs even without a browser trace.
Use `npx --no-install playwright show-trace /absolute/local/path/trace.zip` locally;
never upload traces, generated sites or executables. Teardown removes only owned
temporary trees and closes owned contexts/listeners/children. One worker avoids
clipboard interference; zero retries keeps failures visible.

This implements the 007 MVP and US2 corpus increment. `serve` is covered separately below; PDF, navi and
playgrounds remain deferred. CI runs default and explicit all-corpus suites,
without publishing artifacts or deploying. See `docs/e2e-corpus.md` for pinned
product identities; no sibling executable is needed on PATH. See `docs/e2e-verification.md` for actual results.


## Authenticated Studio journeys

`npm run test:studio` runs the real read-only `smill serve` with installed pinned
Hugo in desktop/mobile contexts. Six scenarios per viewport include saved-file
updates, parser/renderer failures and repair, peers, atomic replacement/removal,
100 rapid-save sequences, same-port restart, disconnect/reconnect, source scope,
inert variants, first-state freshness and isolated failed-state-request recovery.

The server executable is the same owned temporary CLI from the existing tools
fixture. Browser bootstrap is consumed with fixed Node HTTP request headers;
cookies stay in memory and startup stdout is never retained. Browser history and
failure screenshots contain only the clean local origin. Studio configuration
sets tracing **off**, including failures, because authenticated traces can expose
cookies. Every browser request still obeys the exact-origin guard. No artifact
upload or deployment step is added. See `docs/studio-verification.md` and
`docs/studio-security.md`; this suite does not claim T049 performance/usability.

## Asset fingerprints

`assets.spec.mjs` runs six cases: origin-root and `/sheetmill/` mounts in each
browser project. It verifies SHA-256 filenames against actual response bytes,
CSS/script/font resource loading, shared home/sheet URLs and absence of unversioned
asset requests. JavaScript-disabled contexts still load CSS and fonts. These
checks run in both default and all-corpus selections; no deployment is involved.

## Opt-in T049 reference-machine checks

After the same explicit pinned-tool setup, run `npm run test:performance` with
`SHEETMILL_HUGO` set. This launches one desktop Chromium measurement test, with
100 synthetic 8 KiB sheets, 20 fresh-process/fresh-port starts, 100 timed saves
and 100 rapid-save correctness sequences. It enforces SC-002/SC-003 counts and
writes every raw timing sample plus sanitized machine/tool metadata to local
`performance.json` under a unique ignored `.local/e2e/performance/run-*/` directory.
The report identifies the Git revision and measured executable SHA-256; tracked
runtime and harness edits must be committed before the run. No samples are discarded; retries
are disabled. Authenticated traces, screenshots and video are disabled.

Run on an otherwise idle documented workstation. Ordinary CI runs the evaluator
unit tests via `npm run test:harness`; machine-dependent timing thresholds are
opt-in and are not presented as hosted-runner or universal performance guarantees.
The five-author SC-006 usability study is separate and remains pending. See
[Studio verification](../../docs/studio-verification.md) for methods and results.

## Compact code copying

`copy.spec.mjs` verifies that the copy control reserves no header row and overlaps
no code, including a long unbroken token on desktop/mobile. It checks keyboard
activation, exact clipboard contents, fixed-size success/failure feedback and
readable source without JavaScript. It runs in default and all-corpus selections.

Tag/navigation journeys exercise both layouts at root and `/sheetmill/` mounts,
with desktop, mobile and no-JavaScript projects. Independently authored fixtures
verify native category disclosures, current-page state, persistence, exact tag
membership/counts, Unicode/punctuation routes, canonical links and all six palette
choices with light/dark contrast measurements. Two fresh Hugo outputs must match
byte for byte. Commands in the fixtures remain displayed content.
