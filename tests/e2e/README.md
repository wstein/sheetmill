# One-sheet web E2E

Development-only package: Node **24.21.0**, Playwright **1.64.0**, bundled Chromium
revision **1248** (156.0.8078.4), Go **1.27.2** and installed Hugo **0.167.0**.
No sibling CLI, container or account is required. No production JavaScript compiler.

Explicit setup, from this directory:

```sh
npm ci
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

Exactly one real-product corpus case, `smill`, is accepted. Desktop and mobile run
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

This implements only the 007 MVP. `serve`, sibling sheets, PDF, navi, playgrounds
and deployment remain outside it. See `docs/e2e-verification.md` for actual results.
