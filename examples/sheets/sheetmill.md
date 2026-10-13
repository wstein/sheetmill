---
title: SheetMill essentials
category: workflow
tags: [sheetmill, markdown, compiler]
source:
  license: EUPL-1.2
  attribution: SheetMill contributors
  note: Independently authored example; commands are displayed, never executed.
---
Write once, keep your references close. SheetMill turns structured Markdown into a quiet place to find the details.

## The authoring loop {#authoring}
### Check your source {.col-span-2 #check}
Validate a single sheet, or every `.md` file directly inside a directory.

```sh
smill check ./sheets
smill inspect ./sheets/editor.md
```

`inspect` prints the semantic representation as deterministic JSON. A successful `check` stays silent.

### Know the signals {.note #signals}
| Exit | Meaning |
| :-- | :-- |
| `0` | Ready to continue |
| `1` | Fix a source diagnostic |
| `2` | Check arguments or file access |

### Give it structure {#structure}
Use level-two headings for **sections**, and level-three headings for **cards**.

```markdown
## Navigation
### Move around {.col-span-2}
Your reference goes here.
```

### Make a local library {.col-span-2 #local-library}
Export a new project, then invoke your explicitly installed Hugo tool.

```sh
smill build hugo ./sheets --out ./local-project
hugo --source ./local-project --destination ./public --clock 2000-01-01T00:00:00Z --noBuildLock
```

The exporter does not download or launch Hugo. See [links and safety](reference-writing.md#links) for content rules.

## Before you share {#sharing}
### A small checklist {.col-span-2 #checklist}
- [x] Give the sheet a clear title.
- [x] Keep examples short and useful.
- [ ] Check source attribution and licenses.
- [ ] Review the generated page on a small screen.

> A successful import preserves provenance. It does not grant redistribution rights.

### Three useful habits {#habits}
1. Prefer explicit card IDs for stable links.
2. Keep one purpose per card.
3. Commit small, reviewable changes.
