# Review evidence — PR #35

Review baseline: `fdee8ff420368247801e21572653bd9fe8ded8ed`, in isolated worktree
`/private/tmp/sheetmill-pr35-review-20261014`. No merge or deployment.
Concurrent author commits `5815002` and `74f0291` were preserved during rebase,
including their three-column print grid, corner/tint selectors and contrast tests.

## Reproduced findings and fixes

| Finding | Evidence | Fix |
| --- | --- | --- |
| Constrained descriptions stayed left of centered headings | Browser center difference was 304px at 1280px on home/sheet pages | Theme-scoped `margin-inline:auto` |
| Fenced command text retained inline-code tint | Computed code background was `rgb(233, 236, 255)` inside the command row | Reset fenced-code background only; inline code retains its styling |
| Wide table columns collapsed into single-letter text | Mobile screenshot showed vertically broken Meaning header; table did not scroll | Natural table wrapping and internal horizontal scrolling on screen; fixed wrapping in print media |
| New theme absent from browser CI | Existing explicit testMatch covered only other journeys | Add six reference-cards journeys to normal/full-corpus suites |

Each styling failure was observed before its corresponding fix. Sol reviewed the
initial theme and subsequent deltas read-only; it also requested explicit density
state assertions, which were added. A suspected print tint issue was withdrawn:
existing print tokens already make description backgrounds white.

## Actual verification

- Pinned Node 24.21.0 / Playwright 1.64.0 / Chromium 156.0.8078.4 / Hugo 0.167.0.
- Complete reader/corpus browser suite: **51/51 pass**, including all six new theme
  journeys at root/subpath in desktop/mobile/no-JavaScript contexts.
- Widths: 320, 390, 640, 650, 700, 760, 900, 1100 and 1280px; no document overflow,
  badges stay above content, and spanning cards stay within their grid.
- All six palettes in both light/dark modes meet 4.5:1 computed badge contrast;
  no-JavaScript system dark mode is also checked.
- Clipboard text matches literal source, excluding generated `$`; non-shell fences
  do not gain prompts. Sources remain unchanged and external requests are guarded.
- Wide table actually scrolls and exposes its readable far column; print-media
  mode wraps the table within its tile, hides Browse and removes shadows.
- Screenshots inspected for mobile screen and desktop print media. Screenshot
  capture waits for visible content painting only where JavaScript is enabled.
- `SHEETMILL_HUGO=/opt/homebrew/bin/hugo go test ./...`, `go vet ./...` and
  `git diff --check`: pass. Reference-cards Go tests include actual Hugo output.

Exact final-head Sol review and final CI status are posted on the PR. Existing
themes retain their asset isolation and pass their existing browser journeys.

## Limits

Chromium only; no Firefox/WebKit, physical print preview or PDF generation.
An additional old/new site byte-comparison command was rejected by automatic
approval review because it classified its temporary CLI binaries as distribution
builds. That optional command did not run and was not retried through a workaround.
No independent baseline-byte-comparison result is claimed here; the accepted
Go/Hugo and browser checks above provide the review evidence used for these fixes.
