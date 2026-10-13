# Hugo and final brand verification — 2026-10-13

Canonical checkout: `~/github.com/wstein/sheetmill`, branch
`codex/004-hugo-renderer`. Production code independently reviewed at exact
`32a4cb533185e742bf2d1820bd57b7e543b853cb`; subsequent changes only record evidence
and reconcile documentation. [PR #3](https://github.com/wstein/sheetmill/pull/3).

## Actual local checks

Go 1.27.2; Cobra v1.10.2; Hugo 0.167.0 extended (Homebrew); macOS arm64.
Go verification used `GOPROXY=off GOCACHE=/private/tmp/sheetmill-go-cache`.

- `go test ./...` and `go test -race ./...` passed, with
  `SHEETMILL_HUGO=/opt/homebrew/bin/hugo` explicitly enabling actual local generation.
- `go vet ./...`, `go mod verify`, formatting and `git diff --check` passed.
- `python3 -m unittest discover -s assets -p 'test_*.py'` passed all three tests.
- Two independently exported projects, one selected by directory and one by reversed
  explicit file order, produced identical project and public bytes with pinned Hugo
  and `--clock 2000-01-01T00:00:00Z --noBuildLock --quiet`.
- Actual final demo: two sheets, 15 public files. SHA-256 over sorted UTF-8 relative
  filename + NUL + bytes + NUL: `c4f965b4f1c77ce0dfea564df54cb0e13074670d80313abcc979a8f999b6db87`.
- Complete project including public output: 38 files, same digest method:
  `ff001a3b27d8fc84a49dee11e5276a3773f4f5b36fbd91be00ffcebfd43ae974`.

Tests were observed failing before implementing the renderer/CLI, then passing.
Regression tests reproduced review findings before fixes: help-output IO errors,
nested heading/image links, missing diagnostic filenames, dot-relative links,
spans through 12, interior/trailing URL spaces and links inside image alternative
text. Unsafe schemes remain rejected. Original source shortcodes and HTML stay inert.

## Visual and interaction evidence

Local demo served only at `127.0.0.1:1313`, from ignored `.local/hugo-final/public`.
Browser checks covered 1280 px desktop and 390 × 844 mobile. Cards fit in one
mobile column with measured `scrollWidth == clientWidth == 390`; headings, code,
tables, task states and provenance were inspected. Search returned a card, Tab
focused its link and Enter navigated to its anchor. Cross-sheet links reached the
selected destination; copy reported “Copied”; light/dark controls switched the
identity and UI. Screenshots were kept locally, not uploaded or committed.
Navigation/content availability without JavaScript was checked in generated HTML;
no browser JavaScript-disable test was performed. Print CSS was inspected; no PDF
or physical print layout was generated.

The extended brand comparison includes normal/small icons, light/dark/accessible
wordmarks, monochrome, banner and social assets. Sol independently reproduced all
12 SVGs byte-for-byte from pinned fonts/tools, reproduced all five PNGs with
Inkscape 1.4.4, verified the web-font hash and visually inspected the board/exports.
Human visual acceptance of exact masters remains available through the board/PR;
concept approval is not recorded as automatic approval of exact paths.

## Independent review and CI boundary

Sol performed an independent scratch-snapshot review and re-review at the exact
production SHA above. All five renderer findings, two follow-up edge cases and
Cobra help-output finding are closed; no actionable production findings remain.
The final dependency-count documentation correction was also applied.

Hosted CI is reported separately in the PR checks; local passes are not presented
as Linux CI results. CI fetches a SHA-256-pinned official Hugo development archive,
checks actual repeated output, and publishes no site artifacts. This work makes no
full compatibility, dependency vulnerability scan or cross-browser guarantee.

No merge, deployment, release, PDF/navi build or generated-site publication occurred.
