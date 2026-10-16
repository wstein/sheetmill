# Implementation Plan: Sheet freshness

**Branch**: `codex/017-freshness` | **Date**: 2026-10-16 | **Spec**: [spec.md](spec.md)

## Summary

One shared validator (`validate.Freshness`) called from `validate.Metadata` and the
Hugo direct-IR boundary; two optional page fields; one partial rendering the facts in
three layouts; inert Reference-import archival; navi regression pins; docs, schema,
a small golden, and real values on the demo sheets.

## Technical Context

**Language**: Go 1.27.2 (`.go-version`), standard library only (`time`, `unicode`,
`regexp` or manual digit scan). No new dependencies.
**Rendering**: existing Hugo theme under `internal/render/hugo/theme`; installed Hugo
via `SHEETMILL_HUGO=/opt/bundledock/toolbox/bin/hugo`.
**Testing**: negative tables, renderer data/markup tests, real-Hugo clock test,
navi pair test, import tests, parser golden (`-update`, inspected).
**Constraints**: no clock, no network, no JS, no execution; IR 1; no PDF, deployment,
merge or release. Node unavailable: no browser E2E claimed.

## Constitution Check

- 1 Determinism: PASS. Values are authored strings, never parsed into `time` for
  output, no `now`; clock-independence is tested.
- 2 Canonical source: PASS. Public contract goes to `docs/format/gfm-plus.md`, schema
  and diagnostics docs with tests; compatibility note recorded.
- 3 Loss-aware imports: PASS. Archive plus `IMP112`; no inference, no network.
- 4 Inert input: PASS. Text only; escaped.
- 5 Focused Go design: PASS. One validator beside `Navigation`; reuses the existing
  invisible-character predicate pattern.
- 6 Independent product: PASS. No sibling runtime dependency; demo versions come from
  the sheets' own notes.
- 7 Human publication control: PASS. Draft PR authorized 2026-10-16; no merge/deploy/
  build artifacts.
- 8 Review evidence: PASS. Rounds recorded in [analysis.md](analysis.md); actual
  commands and gaps reported.

## Project Structure

```text
specs/017-freshness/{spec,plan,research,data-model,tasks,analysis}.md
specs/017-freshness/contracts/freshness.md
specs/017-freshness/checklists/requirements.md
internal/validate/freshness.go (+_test)       shared rules
internal/validate/validate.go                  call Freshness
internal/render/hugo/hugo.go                   direct-IR check, page fields
internal/render/hugo/freshness_test.go         data + markup + clock tests
internal/render/hugo/theme/layouts/{home,page}.html, partials/freshness.html
internal/render/hugo/theme/static/assets/{site,classic}.css
internal/importer/reference/reference.go       archive + IMP112
internal/render/navi/navi_test.go              pair regression
schema/cheatsheet.schema.json, docs/*, examples/sheets/*.md
testdata/reference-compatible/freshness.md(.json), testdata/import/reference/freshness/
```

## Design notes

- `Freshness` mirrors `Navigation`: tolerant of a non-mapping `cheatsheet`, nil
  locations, sorted diagnostics (updated before tool_version).
- Date check: length 10, bytes 0-3,5-6,8-9 ASCII digits, `-` at 4 and 7, then
  `time.Parse("2006-01-02", s)` and `Year()>=1970` (parse enforces 9999).
- Version check: `utf8.RuneCountInString` 1..64, `strings.TrimSpace(s)==s`, no rune
  satisfying `unicode.IsControl`, `unicode.Is(unicode.Cf, r)`, U+2028, U+2029.
  Also reject filler-only values (FR-003). Invalid UTF-8 cannot occur (source decoding rejects it; direct IR strings are
  checked with `utf8.ValidString`).
- Partial `freshness.html` takes a map-like sheet and emits nothing unless a field
  is set; callers wrap using the same `with` guard so no wrapper leaks. Uses
  `{{ with .tool_version }}` and `{{ with .updated }}`; `datetime` uses the same value.
- CSS: scoped selectors `.library-item .library-fresh`, `.page-heading .sheet-fresh`,
  `.classic .sheet-fresh`, `.classic .index-fresh` (above `.library-item p`; the
  `index-fresh` span is the last child of the row, after `.index-license`, which
  `home.html` emits as `a, .index-desc, .index-count, .index-license`),
  colour `var(--muted)`, 13px; `.classic .index-group li{flex-wrap:wrap}` with
  `.classic .index-fresh{flex-basis:100%}`; `cards.css` resets the basis to auto and
  centres `.reference-cards .sheet-fresh`. Separator is a text span, not CSS content. No print rule.
  `TestPerformanceBudget` (site.css+app.js ~30.9 of 32 KiB) guards size.
- Data path guard: the page fields reach templates only via `data/*.json`; the
  generated `content/<id>/index.md` front matter (`title`, `url`, `sheetid`) must not
  gain `updated`/`date`/`lastmod`, or Hugo would use it as page date (a future value
  with `--clock 2000-01-01` could hide the page). T008 asserts the front matter keys.
- Sitemap: the renderer writes `<loc>` only; no `<lastmod>` from `updated` (FR-009).
- Studio renders through the same templates; `make test-studio` joins the final checks.
- Importer: build a single `cheatsheet` archival map before `validate.Metadata`.
- Template tests that need built pages use real Hugo and are SKIPPED without
  `SHEETMILL_HUGO`; a skip does not count as passing. A template-text test (the
  `TestShelfCardMarkupInTemplate` pattern) runs always. The existing `hugoBuild`
  helpers hard-code the clock, so T012 adds a clock-parameterised variant.
- Risk: the CSS fingerprint changes every theme's asset names; tests that hard-code
  asset names must derive them. Checked at T011.
