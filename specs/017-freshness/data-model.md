# Data model — 017

## Canonical source metadata

```yaml
cheatsheet:
  tool: git              # existing, optional
  tool_version: "2.56.0" # new, optional
  updated: 2026-10-14    # new, optional
```

- `cheatsheet.updated`: string `YYYY-MM-DD`, real calendar date, year 1970–9999.
- `cheatsheet.tool_version`: string, 1–64 code points, single-line, trimmed, no
  `Cc`, `Cf`, `Zl`, `Zp` characters.
- Both optional and independent of each other and of `cheatsheet.tool`. Omitted
  fields stay omitted. Other `cheatsheet.*` keys remain open, preserved metadata.
- IR schema 1; no new typed field on `model.Document`. Values live in
  `metadata.cheatsheet` exactly as authored (strings). YAML timestamp scalars are
  already stored as their source text by `internal/frontmatter`.

## Derived settings

`validate.Freshness(metadata, locations) (FreshnessSettings, []Diagnostic)` with
`FreshnessSettings{Updated, ToolVersion string}` (empty = absent). It never reads a
clock and never reformats; the returned `Updated` is the validated input string.
Diagnostics are `VAL004`; `locations` may be nil (direct IR, document start).
If `cheatsheet` is not a mapping the function returns no diagnostic (`VAL003`
owns that case on the source path; on the Hugo direct-IR path a non-mapping
`cheatsheet` is silently ignored, as for `cheatsheet.tool`, and US2.5 does not cover it).

## Renderer data

`page` gains `Updated string json:"updated,omitempty"` and
`ToolVersion string json:"tool_version,omitempty"`, so `data/library.json`,
`data/shelves.json` members, `data/sheets/<id>.json` and, because `tagCatalog`
embeds `page`, `data/tags/<id>.json` and `data/tagindex.json` gain the keys only
when present (accepted; tag templates ignore them). `data/ir/<id>.json` carries the
metadata as authored. Ordering, related scoring and `ShelfTags` do not read them. Search index
text is unchanged.

## Import provenance

Upstream `cheatsheet.updated` and `cheatsheet.tool_version`, of any shape, move to
`source.original_metadata.cheatsheet.<key>` (merged with an existing `cards`
entry) before `validate.Metadata`, each with an `IMP112` warning at the key span.

## Navi

No change to parsing, hashing or `.cheat` text. The companion `metadata` already
serializes the whole metadata mapping, so the fields appear verbatim there.
