---
title: Writing a good reference
category: authoring
description: Practical advice for writing a reference that answers a task quickly, with literal code, deliberate links and recorded sources.
tags: [writing, markdown, reference]
cheatsheet:
  updated: "2026-10-16"
source:
  license: EUPL-1.2
  attribution: SheetMill contributors
  note: Independently authored reference-writing example.
---
A useful reference answers a specific question at the moment you need it. Put the answer first, then just enough context to use it well.

## Shape the content {#content}
### Start with the task {.col-span-2 #task}
Use a short, descriptive title. Write **“Select a line”** instead of “Selection functionality.” Keep related actions together and give each card one job.

| Prefer | Instead of |
| :-- | :-- |
| A concrete action | An abstract category |
| One small example | Every possible option |
| A stable link | “See above” |

### Keep code literal {#code}
Code blocks stay inert and readable. They are never instructions for the compiler to run.

```sh
printf '%s\n' '<hello>'
```

Use the copy control when you need the exact snippet.

### Link deliberately {.col-span-2 #links}
Use [stable card links](#task), a [selected sheet](sheetmill.md#check), or an explicit HTTPS URL such as [the Go documentation](https://go.dev/doc/).

Unsafe URL schemes are diagnosed. Raw HTML is displayed as text; images are retained as links without fetching them.

### Make the scan easy {#scan}
- **Bold** the important distinction.
- Use `code` for literal values.
- Keep a predictable reading order.
- ~~Delete stale advice~~ and replace it with tested examples.

## Keep it trustworthy {#trust}
### Record the source {.col-span-2 #provenance}
Preserve the source repository, revision, path, license and attribution when importing an existing reference.

> Technical compatibility and permission to redistribute are separate decisions.

### Review in context {#review}
- [x] Read the card without its surrounding discussion.
- [ ] Try narrow and wide browser windows.
- [ ] Follow every local link.
