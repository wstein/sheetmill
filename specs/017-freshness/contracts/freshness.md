# Public contract — 017

```yaml
---
title: Git
cheatsheet:
  tool_version: "2.56.0"
  updated: 2026-10-14
---
```

## Source

- Both fields are optional children of the `cheatsheet` mapping.
- `updated` is date-only ISO 8601: exactly four-digit year, two-digit month and day
  separated by `-`, ASCII digits, a real calendar date, year 1970 to 9999. Quoting
  is optional. Times, zones, `YYYY-M-D`, `YYYYMMDD`, `Oct 14 2026`, numbers,
  Booleans, null and lists are errors.
- `updated` is not compared with the current date. The compiler has no clock; an
  author who writes a future date gets that date. Keep it accurate.
- `tool_version` is free text identifying the version of the tool the sheet
  documents: 1–64 characters, one line, no surrounding whitespace, no control or
  invisible format characters. Write the version only (`3.5.3`), not the tool name.
- Errors use `VAL004` at the key (`cheatsheet.updated` / `cheatsheet.tool_version`),
  or at the document start for direct IR. `cheatsheet` not being a mapping stays `VAL003`.

## Rendering (Hugo, all themes)

Text is always escaped; the date is the only machine-readable element.

| Place | Markup when a field is present |
| --- | --- |
| Editorial home card | `<p class="library-fresh"><span>Version V</span><span>Updated <time datetime="D">D</time></span></p>` after the `Tool:` line |
| Compact/cards index row | `<span class="index-fresh">` with the same two spans after the card count |
| Sheet header, editorial | `<p class="sheet-fresh">` after the description, before tag chips |
| Sheet header, classic-compact and reference-cards | same `<p class="sheet-fresh">` after the description |

Only the present spans are emitted; with neither field nothing is emitted (no
wrapper). Separators are CSS `::before` content, not text. No JavaScript, no title
attribute, no icons. Facts print. Colors are existing palette tokens.

## Other contracts

- `smill build navi`: `.cheat` bytes and identities unchanged; companion metadata
  carries `cheatsheet.updated`/`tool_version` verbatim; no new navi tag or diagnostic.
- Reference import: `IMP112` archives upstream values; nothing is inferred.
- Sorting, routes, related references, tag pages, search and the IR schema number
  are unaffected.

## Changed generated output (only for sheets carrying the fields)

`data/library.json`, `data/shelves.json`, `data/sheets.json` (keys `updated`,
`tool_version`); `index.html` (home) and each `<id>/index.html`; `assets/site.css`
and `assets/classic.css` (fingerprinted names change; contents add `.library-fresh`,
`.sheet-fresh`, `.index-fresh` rules); `assets/cards.css` only if its theme needs
overrides. Parser golden `testdata/reference-compatible/freshness.md.json` is new;
no existing golden changes. The navi companion of the demo sheets changes (new
metadata only), their `.cheat` files do not.

## Compatibility

A pre-existing `cheatsheet.updated` or `cheatsheet.tool_version` with another shape
formerly passed as generic metadata and is now `VAL004`. Rename it to an
author-owned key. IR version stays 1.
