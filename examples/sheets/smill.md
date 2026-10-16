---
title: SheetMill CLI
category: reference
description: Check, inspect, export, import and preview sheets with the smill command line, from a first validation to a local Hugo library.
tags: [sheetmill, smill, markdown]
source:
  repository: https://github.com/wstein/sheetmill
  revision: 92d64e3965454540e6b7dab2177edc279ed10147
  path: internal/cli/cli.go
  license: EUPL-1.2
  attribution: SheetMill contributors
  note: Independently authored against implemented CLI commands; examples are inert.
---
Keep the authoring loop small: validate a source, inspect its structure, then export a local reference library.

## Everyday workflow {#workflow}
### Build a local library {.col-span-2 #build-library}
Check the selected sources before exporting a **new** Hugo project.

```sh
smill check ./sheets
smill build hugo ./sheets --out ./new-project
```

The export command does not launch or install Hugo. Generate the website separately with pinned Hugo 0.167.0; serve the resulting public directory at the origin root.

### Check a sheet {#check-sheet}
```sh
smill check ./sheets/editor.md
```

Successful validation stays silent. Diagnostics identify the source, line, column and stable code. See [exit status](#exit-status).

### Inspect semantic IR {#inspect-ir}
```sh
smill inspect ./sheets/editor.md
```

Print the versioned semantic representation as JSON for exactly one valid file. Source metadata and literal code remain inspectable.

### Version and help {#version-help}
```sh
smill version
smill --help
smill build hugo --help
```

The development version prints `smill dev`. Command-specific help documents the implemented flags.

## Inputs and diagnostics {#inputs}
### Choose explicit inputs {.col-span-2 #choose-inputs}
A directory selects lowercase `.md` files **one level deep**. Subdirectories are ignored. Explicit duplicate paths are deduplicated. Selected files and ancestor paths must not be symlinks.

Use `/private/tmp` for macOS scratch sources. Each input is limited to 8 MiB; there is no recursive or stdin mode.

### Exit status {#exit-status}
These statuses describe batch validation and export commands.

| Status | Meaning |
| :-- | :-- |
| `0` | Success |
| `1` | Invalid source or rendering diagnostic |
| `2` | Usage or file access/output failure |

`serve` keeps invalid sources in a running diagnostic session. Normal interruption exits `0`; setup or fatal runtime failures exit `2`.

IO/usage failures take precedence over invalid-source status. Existing export directories are refused, never overwritten.

## Controlled conversion {#conversion}
### Import with provenance {.col-span-2 #import-provenance}
Use `smill import reference --help` to supply the source repository, path, full revision, license and attribution. Select `--dry-run` to inspect conversion without writing; `--report json` prints a deterministic loss report.

A written `--out` target must be a new `.md` file. License metadata preserves provenance; it does not grant permission to redistribute imported material.

### Keep examples inert {#inert-examples}
- [x] Treat code blocks as reference content.
- [x] Keep IDs stable for [library links](#build-library).
- [ ] Review a generated page before sharing it.

Raw HTML is shown escaped and images remain unfetched links with warnings. Source shortcodes, scripts and shell snippets are never executed by the compiler.


## Local live preview {#preview}
### Preview saved files {.col-span-2 #serve-project}
```sh
smill serve ./sheets --hugo /absolute/path/to/hugo
smill serve ./sheets/editor.md --open --port 0
```

Use installed Hugo 0.167.0. The default selection is the current directory; a file selects only that file. `--port 0` chooses an available IPv4 loopback port. `--open` opens the browser after readiness; a launch failure leaves the printed address available.

The printed bootstrap address is private and usable once. The browser is read-only: save Markdown in your external editor. Pending, invalid, removed and disconnected states are explicit; an invalid save may retain a labeled last-valid preview. No source snippet, PDF, navi command or deployment runs. Interrupt the CLI to close the session and remove its owned temporary output.
