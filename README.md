# SheetMill

![SheetMill — One source. Every reference.](assets/banner.svg)

[![CI](https://github.com/wstein/sheetmill/actions/workflows/check.yml/badge.svg)](https://github.com/wstein/sheetmill/actions/workflows/check.yml)
[![Go version](https://img.shields.io/badge/Go-1.27.2-00ADD8?logo=go&logoColor=white)](go.mod)
[![License: EUPL-1.2](https://img.shields.io/badge/license-EUPL--1.2-blue)](LICENSE)

Git-native cheatsheet compiler. Turn Markdown and YAML frontmatter into beautiful
web references, Typst PDFs, and interactive navi commands.

**Status:** foundation, parser, validation, deterministic JSON inspection and
controlled Reference import and self-contained Hugo project export are implemented.
Local website generation and read-only live preview are supported with pinned
Hugo 0.167.0. `smill build navi` exports explicitly marked command cards as
navi `.cheat` files (text only; navi is never run). PDF output
remains planned for a later milestone.
Read the [SheetMill manual](https://wstein.github.io/sheetmill/), itself built
with SheetMill from `docs/`, and try the [live demo library](https://wstein.github.io/sheetmill/demo/),
generated from the independently authored example sheets. After the approved merge
and deployment, the same sheets are also published in the
[compact](https://wstein.github.io/sheetmill/demo-compact/) and
[cards](https://wstein.github.io/sheetmill/demo-cards/) layouts.
Nothing in a source sheet is executed.

The CLI executable is `smill`. Go **1.27.2** is the pinned development toolchain.
There are no published binaries or install packages yet.

```text
smill version
smill check sheet.md
smill check ./sheets
smill inspect sheet.md
smill build hugo ./sheets --out ./new-project
smill serve ./sheets --open --hugo /absolute/path/to/hugo
```

`check` takes explicit files or one-level directories; `inspect` prints versioned
JSON for exactly one file. `serve` watches externally saved files on IPv4
loopback, with a private one-use bootstrap address and no browser write controls.
See [Studio security and usage](docs/studio-security.md). Diagnostics go to stderr. See [CLI](docs/cli.md).

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

AI agents are first-class authors: the [agent guide](docs/agent-guide.md) covers
setup and the authoring loop, and [llms.txt](llms.txt) indexes resources for
debugging. `smill --help` points agents to both.

Convert explicit external files using the [Reference import adapter](docs/imports.md),
with provenance and deterministic loss reports. It never overwrites an existing file.

Read the [source format](docs/format/gfm-plus.md), [compatibility boundary](docs/format/compatibility.md),
[architecture](docs/architecture.md), and [three branding proposals](docs/brand-options.md).
The selected direction is Editorial Stack with coral/periwinkle colors; see the
[brand guide and finalized masters](assets/BRAND.md), with an extended
[visual review board](assets/preview.svg). See [rated recommendations](docs/recommendations.md).

## Development

```sh
make             # Show available targets
make fmt         # Format Go sources
make check       # Formatting check, Go tests, vet and whitespace
make race        # Race tests
```

CI uses the same Makefile targets. `make check` requires Go, Git and Make;
Hugo and browser checks are separate. After installing the pinned development
tools described in [the browser harness guide](tests/e2e/README.md):

```sh
export SHEETMILL_HUGO=/absolute/path/to/hugo
make test-hugo
make test-harness
make test-browser-corpus
make test-browser-studio
```

Targets do not install dependencies or regenerate goldens. Tests compile temporary
test executables; no distribution binary is produced. There is no `make build`
or deployment target in this increment.
[Local Hugo generation](docs/hugo.md) describes the explicit export/build workflow.
The CLI uses Cobra **v1.10.2**. Goldmark supplies GFM parsing and the YAML organization supplies YAML decoding;
[dependency and license provenance](docs/provenance.md) explains these choices.

SheetMill has no runtime dependency on BundleDock, WorkHarbor or CrewBook.
The numbered [feature specifications](specs/README.md) track implementation and future work.

Licensed under the [European Union Public Licence 1.2](LICENSE).
