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
| `cheatsheet` | Optional mapping |
| `cheatsheet.schema` | Optional integer, exactly `1`; omitted means current schema |
| `cheatsheet.tool` | Optional nonempty string |

All other JSON-compatible fields, including nested `cheatsheet` fields, are
preserved in `metadata`. `cheatsheet.navi` is generic metadata to the parser and
`smill check`; only `smill build navi` validates it, under the
[navi export contract](navi-profile.md).
The [JSON Schema](../../schema/cheatsheet.schema.json) describes
metadata after YAML decoding; parser limits and duplicate keys are additional rules.

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
| `{.navi}` | Marks a command card for `smill build navi`, which also needs an explicit `{#id}`; see [navi export](navi-profile.md) |
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
