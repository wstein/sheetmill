# Public contract — 017

```yaml
---
title: Git
cheatsheet:
  tool_version: "2.56.0"
  updated: "2026-10-14"
---
```

## Source

- Both fields are optional children of the `cheatsheet` mapping.
- `updated` is date-only ISO 8601: exactly four-digit year, two-digit month and day
  separated by `-`, ASCII digits, a real calendar date, year 1970 to 9999. Quoting
  is optional for SheetMill, but quote it for YAML 1.1 tools that turn bare dates
  into date objects. Times, zones, `YYYY-M-D`, `YYYYMMDD`, `Oct 14 2026`, numbers,
  Booleans, null and lists are errors.
- `updated` is not compared with the current date. The compiler has no clock; an
  author who writes a future date gets that date. Keep it accurate.
- `tool_version` is free text identifying the version of the tool the sheet
  documents: 1–64 characters, one line, no surrounding whitespace, no control or
  invisible format characters, not only blank-looking filler characters (U+115F,
  U+1160, U+3164, U+FFA0, U+2800); other unusual marks are not policed. Write the version only (`3.5.3`), not the tool name; so omit the field where the bare
  number would be ambiguous without it (the demo `tar` sheet documents bsdtar 3.5.3,
  while GNU tar is 1.x; see research.md). Freshness is read from the validated
  metadata and exposed through `data/*.json` only, never through content front matter.
- Quote `tool_version`: an unquoted `3.81` or `3.10` is a YAML number, not text, and is
  rejected with its own `VAL004` message, "must be a quoted YAML string", rather than
  the length and character rule. The same message covers a non-string `updated`.
- Errors use `VAL004` at the key (`cheatsheet.updated` / `cheatsheet.tool_version`),
  or at the document start for direct IR. `cheatsheet` not being a mapping stays `VAL003`.

## Rendering (Hugo, all themes)

Text is always escaped; the date is the only machine-readable element.

| Place | Markup when a field is present |
| --- | --- |
| Editorial home card | `<p class="library-fresh"><span>Version V</span> <span aria-hidden="true">·</span> <span>Updated <time datetime="D">D</time></span></p>` after the optional `.library-tool` and before `.library-description`, the same place whether or not `tool` is set (no demo sheet has `cheatsheet.tool` and FR-013 forbids adding it) (separator only between two present facts) |
| Compact/cards index row | `<span class="index-fresh">` with the same two spans as the **last child** of the row, after `.index-desc`, `.index-count` and `.index-license` (the template order; all 15 demo sheets carry a license, so placing it earlier would strand the license after a full-width line) |
| Sheet header, editorial | `<p class="sheet-fresh">` after the description, before tag chips |
| Sheet header, classic-compact and reference-cards | same `<p class="sheet-fresh">` placed directly after `</div>` of `.sheet-head` and after the optional `<p class="sheet-description">` (the description sits outside `.sheet-head`); with no description it follows `.sheet-head` directly |

Only the present spans are emitted; with neither field nothing is emitted (no
wrapper). The separator is a real aria-hidden `·` span with the spaces outside it, never CSS content. No
JavaScript, no title attribute, no icons. Facts print. Colors are existing palette
tokens. CSS selectors are scoped (`.library-item .library-fresh`,
`.page-heading .sheet-fresh`, `.classic .sheet-fresh`) so they out-rank
`.library-item p`. In the classic layouts `.classic .index-group li:has(.index-fresh){flex-wrap:wrap}` is added (scoped with `:has()` so
rows without facts keep their wrapping; the theme already relies on `:has()` in `site.css` and
`cards.css`, supported by current Chrome, Safari and Firefox, and where it is unsupported the
rule is dropped and the facts stay on the row's single line) and
`.classic .index-fresh{flex-basis:100%}` gives the facts their own last line; `cards.css`
resets it (`.reference-cards .index-fresh{flex-basis:auto}`, its rows are columns)
and centres `.reference-cards .sheet-fresh` like the description. The desktop
layout of these rules is unverified until a browser check is run. Tag pages, the
tag index, the sidebar and search results do not show freshness.

## Other contracts

- `smill build navi`: `.cheat` bytes and identities unchanged; companion metadata
  carries `cheatsheet.updated`/`tool_version` verbatim; no new navi tag or navi code.
  Source files pass the shared parser first, so invalid freshness fails with `VAL004`
  (also `smill check`, Studio); direct-IR navi input is not validated.
- Reference import: `IMP112` archives upstream values; nothing is inferred.
- Sorting, routes, related references, tag pages, search and the IR schema number
  are unaffected.

## Changed generated output (only for sheets carrying the fields)

For sheets carrying the fields: `data/library.json`, `data/shelves.json`,
`data/sheets/<id>.json`, `data/tags/<id>.json`, `data/tagindex.json` (keys `updated`,
`tool_version`), `data/ir/<id>.json` (metadata), `index.html` and each
`<id>/index.html`. For every build: the fingerprinted name of each emitted stylesheet (editorial emits
`assets/site.css` only; `classic-compact` adds `assets/classic.css`; `reference-cards`
also `assets/cards.css`; `app.js` is unchanged) (contents gain the `.library-fresh`,
`.sheet-fresh`, `.index-fresh` rules; `cards.css` the centring rule), hence every
href in `baseof.html` output and `hugo.json` params. The Studio preview renders through the same templates
(`internal/studio/compile.go` calls `hugo.Render`), so it shows the facts for sheets
that carry them and its stylesheet name changes too (`make test-studio`). Parser golden `testdata/reference-compatible/freshness.md.json` is new;
no existing golden changes. The navi companion of the demo sheets changes (new
metadata only), their `.cheat` files do not.

## Compatibility

A pre-existing `cheatsheet.updated` or `cheatsheet.tool_version` with another shape
formerly passed as generic metadata and is now `VAL004`. Rename it to an
author-owned key. IR version stays 1.
