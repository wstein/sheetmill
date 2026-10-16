---
title: "Source format v0.1"
description: "Frontmatter, sections, cards, heading attributes, IDs and the JSON IR."
category: reference
---
Input is a UTF-8, lowercase `.md` file, at most 8 MiB. No BOM or text may precede
frontmatter. LF and CRLF are accepted and preserved. The opening and closing
YAML delimiters must each be exactly `---` on their own line; the opening needs
a newline. No evaluation, includes, fetching or snippet execution occurs.

## Frontmatter

The YAML root must be one mapping. Keys at every depth must be strings and unique.
Anchors, aliases, merge keys, custom tags, nonfinite numbers and nesting beyond
64 levels are errors. Supported values are strings, booleans, null, signed 64-bit
integers, finite floating-point values, arrays and string-keyed mappings.
YAML timestamps are preserved as strings. Explicit null remains null.
Known fields use strict types without implicit scalar/array coercion:

| Field | Rule |
| --- | --- |
| `title` | Required nonempty string |
| `category`, `description` | Optional strings |
| `tags` | Optional array of nonempty strings |
| `navigation` | Optional closed mapping: [library placement](#library-placement) |
| `cheatsheet` | Optional mapping |
| `cheatsheet.schema` | Optional integer, exactly `1`; omitted means current schema |
| `cheatsheet.tool` | Optional nonempty string |
| `cheatsheet.updated` | Optional date `YYYY-MM-DD`: [sheet freshness](#sheet-freshness) |
| `cheatsheet.tool_version` | Optional one-line string, 1 to 64 characters: [sheet freshness](#sheet-freshness) |
| `cheatsheet.cards` | Optional shared card/parameter mapping; see [command cards](command-cards.md) |

All other JSON-compatible fields, including unreserved nested `cheatsheet` fields, are
preserved in `metadata`. `cheatsheet.navi` is generic metadata to the parser and
`smill check`; only `smill build navi` validates it, under the
[navi export contract](navi-profile.md).
The [JSON Schema](../../schema/cheatsheet.schema.json) describes
metadata after YAML decoding; parser limits and duplicate keys are additional rules.

## Shared command cards {#shared-command-cards}

Use `{#id .command}` for target-neutral command semantics and
`cheatsheet.cards.<id>.params` for card-local parameter declarations. The
[shared command-card contract](command-cards.md) specifies core validation,
migration and exact navi output/identity preservation. Ordinary shell code blocks
are not automatically exported. Legacy `.navi` and `cheatsheet.navi.cards` remain
supported aliases; conflicting same-card declarations are rejected, never merged
partially or silently overridden. No commands or providers execute.

## Library placement {#library-placement}

`navigation` places a sheet in the Hugo library sidebar and home page without
changing its category, tags, URL or relevance:

```yaml
navigation:
  group: SheetMill
  order: 20
```

- `navigation` is optional; when present it must be a mapping, and `{}` is valid.
  Only `group` and `order` are allowed. A misspelling, or a deferred field such as
  `title`, fails with `VAL003`.
- `group` is a string with non-whitespace content, kept exactly as written. When
  omitted, the sheet's `category` names its group, then Other references. A
  supplied but invalid group never falls back silently.
- `order` is an integer from 0 to 1,000,000. Strings, Booleans, null and floats
  (including `1.0`) are errors; leaving it out differs from an explicit `0`.
- Within a group, ordered sheets come first, ascending. Remaining sheets and equal
  orders compare by Unicode-lowercased title, then the original title bytes, then
  route. Equal orders are allowed and never warn. Use steps of ten to leave room.
- Groups still sort by name, with the ungrouped shelf last. A group named
  `Other references` stays a named group, separate from the fallback.
- Only library placement changes. Related sheets, search, tag pages, URLs and
  cards keep their existing behavior.

Navigation was previously ordinary preserved metadata. It is now a reserved,
closed namespace, so an existing `navigation` object with other keys must be
renamed to an author-owned key or brought into this contract, even though the IR
schema stays at 1. The JSON Schema describes metadata after YAML decoding; the
parser's strict scalar typing is what rejects `order: 1.0`, which JSON Schema
alone cannot tell apart from `1`. Reference imports archive any upstream
`navigation` as inert provenance (`IMP102`, see [imports](../imports.md)); the
navi companion preserves it verbatim without changing `.cheat` output.

## Sheet freshness {#sheet-freshness}

`cheatsheet.updated` and `cheatsheet.tool_version` state when the sheet was last
updated and which version of the tool it documents. The Hugo renderer shows them on
library cards and sheet headers; nothing else reads them.

```yaml
cheatsheet:
  tool_version: "2.56.0"
  updated: "2026-10-14"
```

- Both fields are optional and independent. Nothing is defaulted, inferred or
  written back. `cheatsheet.tool` stays the tool name.
- `updated` is a date written exactly `YYYY-MM-DD` with ASCII digits that names a
  real calendar date (leap days included) from 1970 to 9999. Times, zones, other
  spellings (`2026-1-5`, `20261016`, `16 Oct 2026`), numbers, Booleans, null and
  lists are errors. Quote the date: YAML 1.1 tools turn a bare date into a date
  object. SheetMill accepts both, because it keeps YAML timestamps as strings.
- It is not compared with the current date, since the compiler has no clock. A
  future date is shown as written, so keep it accurate and set it only from
  verified evidence, never by guessing.
- `tool_version` is free text shown as written (`10.3p1`, `2.6.0-FreeBSD`), with no
  `v` prefix or semantic-version parsing. Write the version only, not the tool
  name, and leave it out where the bare number would be ambiguous. It must be one
  line of 1 to 64 characters, without surrounding whitespace, control or invisible
  format characters, and not only blank-looking fillers (U+115F, U+1160, U+3164,
  U+FFA0, U+2800). Private-use and unassigned characters are accepted. Quote the
  version: an unquoted `3.81` or `3.10` is a YAML number (and `3.10` loses its
  zero), so it fails with `VAL004` asking for a quoted string, e.g. `"3.81"`.
- Invalid values fail with `VAL004` at the key, in `smill check`, `smill build navi`
  and the preview, and for semantic input given directly to the Hugo renderer.
  `cheatsheet` that is not a mapping stays `VAL003`.
- Hugo shows `Version V` and `Updated D` as escaped text, the date inside
  `<time datetime="D">`, on editorial cards, compact index rows and sheet headers.
  Ordering, URLs, related sheets, tags and search do not use them, and the output
  never depends on the build date. The sitemap does not get a `<lastmod>`.
- `smill build navi` ignores them: `.cheat` bytes and card identities do not
  change, the version is not a navi tag, and the metadata companion carries them as
  ordinary metadata. Reference imports archive upstream values with `IMP112`
  instead of activating them ([imports](../imports.md)).

Before this feature, a `cheatsheet.updated` or `cheatsheet.tool_version` of another
shape was ordinary preserved metadata. Rename such a value to an author-owned key; the
IR schema stays at 1.

## Markdown and cards

Goldmark's GFM extension supplies CommonMark blocks and inline content, tables,
strikethrough, task lists and autolinks. Root-level `##` headings start sections;
root-level `###` headings start cards and require a preceding section. `#` and
`####` through `######` remain ordinary content. Headings inside quotes, lists,
or code never start sections/cards. Setext level-two headings also start sections.
A document may have no sections; its ordinary content stays in `preamble`.

Content before the first section is preamble. Content after a section heading
and before its first card belongs to that section. A card ends at the next
root-level section/card or EOF. Order and nesting are retained. Code block text
retains literal bytes after CommonMark's container indentation handling, including
newlines and shorter fences. The complete original Markdown body is also stored.
Email autolink destinations include the `mailto:` scheme so the IR preserves
mail-link semantics independently of Goldmark. Ordinary URL destinations keep
their schemes. Raw HTML is inert data; no renderer exists. Future renderers must
sanitize it.

## Heading attributes

A section/card heading may end with one plain-text attribute group. Attributes
inside inline code/emphasis are ordinary content. The structural `title` excludes
the attribute group and decodes punctuation escapes; `heading` retains original
inline nodes for source fidelity. An escaped opening brace remains literal text;
only an unescaped trailing group defines attributes.
Nested content headings do not interpret attributes. Attribute groups anywhere
else fail with `ATTR003` instead of silently becoming text: a line holding only a
group (for example `{.marker-none}` after a list or table) or a group ending
any other heading. Escape the brace (`\{.x}`) or use a code span to show one.

| Attribute | Meaning |
| --- | --- |
| `{#id}` | Explicit globally unique ID; ASCII letter then letters, digits, `_`, `-` |
| `{.cols-N}` | Section grid hint, decimal `N` from 1 through 6 (default 3); sections only |
| `{.col-span-N}` | Card web-layout hint, decimal `N` from 1 through 12 |
| `{.row-span-N}` | Card web-layout hint, decimal `N` from 1 through 12 |
| `{.note}`, `{.tip}`, `{.warning}` | SheetMill semantic card classes |
| `{.keys}` | Key table card: inline code in a table's first column renders as keys (`Ctrl+C` shows two keys) |
| `{.command}` | Shared command-card semantics; requires explicit `{#id}`; see [command cards](command-cards.md) |
| `{.navi}` | Legacy command selection alias; existing navi warning/skip behavior is retained; see [navi export](navi-profile.md) |
| `{.wrap}` | Inert legacy compatibility hint; does not change semantic content |

Combine tokens inside one group, for example `{.col-span-2 .note #example}`.
Repeated/conflicting tokens, unsupported attributes and malformed groups fail.
Sections allow only `#id` and `.cols-N`; cards never span more columns than their
section provides, and narrow screens reduce every grid to two and then one column. No arbitrary CSS, Tailwind rules or theme classes enter
layout semantics. Layout hints never dictate future PDF pagination.

## IDs, spans and JSON

IDs use explicit `#id` when supplied. Otherwise lowercase title letters/digits
are joined by `-`; an empty slug becomes `content`. Duplicates get `-2`, `-3`, etc.
Names are globally reserved in source order, including explicit IDs; a later
explicit ID colliding with a generated ID fails. IDs survive changes to whitespace
or body text but titles/order can change generated IDs. Use explicit IDs for links
that must survive title edits. A card also carries `id_explicit: true` in schema-1
JSON when its heading declares `{#id}`; the field is omitted for generated IDs.
Renderers use this parser-recorded fact without guessing from heading text.

JSON `schema: 1` is the IR version, independent of future source-format versions.
Output has fixed struct field order, lexically sorted map keys, two-space indent,
Go JSON escaping and one trailing LF. Lists representing document structure use
empty arrays. A section's `columns` field appears only when `.cols-N` is given. Optional node properties/children/text are omitted when empty.
No absolute filename, timestamp, host metadata or random ID enters this output.
Identical source bytes produce identical JSON; different line endings are not
normalized and may produce different output.

Spans are half-open byte ranges in the **entire original input**, with zero-based
`offset` and one-based line/rune column. Section/card spans include their heading
and extend to the next boundary. AST node spans anchor original semantic content;
inline spans need not include every Markdown delimiter/destination. The original
body plus its frontmatter-derived offset is the lossless source of syntax.
