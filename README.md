# SheetMill

[![CI](https://github.com/wstein/sheetmill/actions/workflows/check.yml/badge.svg)](https://github.com/wstein/sheetmill/actions/workflows/check.yml)
[![Go version](https://img.shields.io/badge/Go-1.27.2-00ADD8?logo=go&logoColor=white)](go.mod)
[![License: EUPL-1.2](https://img.shields.io/badge/license-EUPL--1.2-blue)](LICENSE)

Git-native cheatsheet compiler. Turn Markdown and YAML frontmatter into beautiful
web references, Typst PDFs, and interactive navi commands.

**Status:** foundation, parser, validation, deterministic JSON inspection and
controlled Reference import are implemented. Web, PDF and navi output are planned
for later milestones.
Nothing in a source sheet is executed.

The CLI executable is `smill`. Go **1.27.2** is the pinned development toolchain.
There are no published binaries or install packages yet.

```text
smill version
smill check sheet.md
smill check ./sheets
smill inspect sheet.md
```

`check` takes explicit files or one-level directories; `inspect` prints versioned
JSON for exactly one file. Diagnostics go to stderr. See [CLI](docs/cli.md).

## Source

````markdown
---
title: Tiny Editor
category: editor
tags: [editor, keyboard]
cheatsheet:
  schema: 1
  tool: tiny
---
## Editing
### Selection {.col-span-2}
| Key | Meaning |
| --- | --- |
| `x` | Select line |
````

Convert explicit external files using the [Reference import adapter](docs/imports.md),
with provenance and deterministic loss reports. It never overwrites an existing file.

Read the [source format](docs/format/gfm-plus.md), [compatibility boundary](docs/format/compatibility.md),
[architecture](docs/architecture.md), and [three branding proposals](docs/brand-options.md).
The palette is still undecided. See [rated recommendations](docs/recommendations.md).

## Development

```sh
gofmt -w cmd internal
go test ./...
go vet ./...
```

Tests compile temporary test executables; no distribution binary is produced.
Goldmark supplies GFM parsing and the YAML organization supplies YAML decoding;
[dependency and license provenance](docs/provenance.md) explains these choices.

SheetMill has no runtime dependency on BundleDock, WorkHarbor or CrewBook.
The numbered [feature specifications](specs/README.md) track implementation and future work.

Licensed under the [European Union Public Licence 1.2](LICENSE).
