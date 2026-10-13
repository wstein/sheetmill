---
title: "Getting started"
description: "What SheetMill is, how to build it and how to write, check and publish your first sheet."
category: guide
---
SheetMill turns Markdown reference sheets with YAML frontmatter into a checked,
structured representation and a fast, self-contained reference website. The
command-line tool is `smill`. Nothing in a sheet is executed or fetched.

See what it produces in the [demo library](https://wstein.github.io/sheetmill/demo/),
built from the example sheets in the repository.

## Install {#install}
### Build from source {.col-span-2 #build}
There are no published binaries yet. Build with Go 1.27.2:

```sh
git clone https://github.com/wstein/sheetmill
cd sheetmill
go build -o ./smill ./cmd/smill
./smill version
```

### Optional tools {#tools}
Building a website and the live preview need Hugo **0.167.0** (exact version).
Checking and inspecting sheets need nothing else.

## Your first sheet {#first-sheet}
### Write it {.col-span-2 #write}
Create `sheets/editor.md`:

````markdown
---
title: Tiny Editor
category: editor
tags: [editor, keyboard]
---
## Editing {#editing}
### Select a line {#select-line}
| Key | Meaning |
| --- | --- |
| `x` | Select line |
````

`##` starts a section and `###` starts a card. The [source format](format/gfm-plus.md#heading-attributes)
lists every rule, including layout hints such as `{.col-span-2}`.

### Check it {#check}
```sh
./smill check sheets
./smill inspect sheets/editor.md
```

`check` is silent when the sheet is valid. Otherwise every message names the file,
line, column and a stable [diagnostic code](cli.md).

## Publish {#publish}
### Build a website {.col-span-2 #website}
```sh
./smill build hugo sheets --out ./site
hugo --source ./site --destination public
```

The export goes into a new directory and is never overwritten. See
[Hugo website export](hugo.md) for hosting under a path such as `/docs/`.

### Preview while editing {#preview}
```sh
./smill serve sheets --open
```

A read-only [live preview](studio-security.md) rebuilds when you save.

## Going further {#further}
### Import and automation {.col-span-2 #automation}
Convert existing Reference sheets with recorded provenance using the
[controlled importer](imports.md). AI agents can follow the [agent guide](agent-guide.md),
and `smill --help` points them to it.
