---
title: Markdown
description: GitHub Flavored Markdown as SheetMill accepts it, plus the frontmatter, section, card and attribute rules of SheetMill sheets.
category: markup
tags: [markdown, gfm, sheetmill, writing]
cheatsheet:
  updated: "2026-10-14"
source:
  license: EUPL-1.2
  attribution: SheetMill contributors
  note: Independently authored from the SheetMill source format v0.1 documentation (docs/format/gfm-plus.md); examples are displayed, never executed.
---
SheetMill reads CommonMark with the GitHub Flavored Markdown extensions for tables, strikethrough, task lists and autolinks. Every example below is Markdown source, shown as code.

## Inline text {.cols-2 #inline}
### Emphasis {#emphasis}
```markdown
*italic* or _italic_
**bold** or __bold__
***bold and italic***
~~struck through~~
```

### Code spans {#code-spans}
```markdown
Run `smill check` before review.
Use ``a `tick` inside`` with a longer run of backticks.
```

Text inside a code span is literal, including `*`, `_` and braces.

### Escaping {#escaping}
```markdown
\*not italic\*
1986\. A year, not a list item
\# not a heading
```

A backslash before ASCII punctuation makes it a literal character.

### Line breaks {#breaks}
```markdown
First paragraph line
continues the same paragraph.

A blank line starts a new paragraph.
End a line with a backslash\
to force a hard break.
```

## Blocks {.cols-2 #blocks}
### Headings {#headings}
```markdown
# Level one
## Level two
### Level three
#### Level four
```

In a sheet, root-level `##` starts a section and `###` starts a card; other levels stay ordinary content.

### Lists {#lists}
```markdown
- milk
- bread
  - whole grain
  - rye

1. preheat
2. bake
3. cool
```

Indent nested items to line up with the parent item's text.

### Task lists {#tasks}
```markdown
- [x] write the sheet
- [x] run smill check
- [ ] ask for review
```

### Quotes and rules {#quotes}
```markdown
> Quoted text can span
> several lines.
>
> > And nest.

---
```

Three or more `-`, `*` or `_` draw a thematic break. Keep a blank line above `---`: directly under text it makes a level-two heading, which starts a new section.

## Links {.cols-2 #links}
### Inline and reference links {.col-span-2 #link-forms}
```markdown
Read the [format notes](https://example.org/format "Optional title").
See the [jq sheet](jq.md#construct) or [this card](#tables).

Reuse a target with [a reference][fmt] or just [fmt].

[fmt]: https://example.org/format
```

Use `other.md#id` to link a card on another sheet and `#id` within the same sheet.

### Autolinks {#autolinks}
```markdown
<https://example.org/docs>
https://example.org/plain
www.example.org
help@example.org
```

GFM links bare URLs and addresses too; email links keep a `mailto:` destination.

### Safe destinations {.warning #safe-links}
| Destination | Hugo renderer |
| :-- | :-- |
| `https://`, `http://` | Accepted |
| `mailto:` | Accepted |
| `other.md#id`, `#id` | Accepted, relative |
| other schemes | Rejected as unsafe |

Raw HTML is kept as inert text and images become plain links; neither is fetched.

## Code and tables {.cols-2 #code-tables}
### Fenced code blocks {.col-span-2 #fences}
````markdown
```sh
smill check sheets/
```

~~~json
{"title": "demo"}
~~~
````

Put a language name after the opening fence. To show a fence inside a fence, make the outer fence longer, as this card does.

### Tables {#tables}
```markdown
| Key | Action | Count |
| :-- | :-: | --: |
| `q` | quit | 1 |
| `w` | write | 12 |
```

Colons in the delimiter row align columns left, centre or right.

### Pipes inside cells {.tip #table-pipes}
```markdown
| Shell | Meaning |
| :-- | :-- |
| `a \| b` | pipe output |
```

Escape a literal `|` in a cell with a backslash, even inside a code span.

## SheetMill structure {.cols-2 #sheet-structure}
### Frontmatter {#frontmatter}
```markdown
---
title: Tide tables
description: Reading tide charts at a glance.
category: outdoors
tags: [tides, sailing]
---
```

`title` is required; unknown fields are kept as metadata, not dropped.

### Sections and cards {.col-span-2 #sections-cards}
```markdown
Short intro before the first section.

## Reading charts {.cols-2 #charts}
### High water {#high-water}
Card content.

### Spring tides {.col-span-2 .tip #spring}
More card content.
```

Sections accept `#id` and `.cols-N` (1 to 6). Cards accept `#id`, `.col-span-N`, `.row-span-N`, `.note`, `.tip` and `.warning`.

### Attribute placement {.warning #attribute-errors}
```markdown
- item
{.compact}

#### Detail {.note}
```

Both lines fail with `ATTR003`: groups are only allowed at the end of root `##` and `###` headings. Write `\{.x}` or a code span to show braces as text.

### Stable IDs {.note #ids}
```markdown
## Install {#install}
### From source {#from-source}
```

Without `#id`, SheetMill derives one from the title, so renaming a heading changes its link. IDs must be unique in the sheet.
