---
title: SheetMill CLI
category: reference
tags: [sheetmill, smill, markdown]
source:
  repository: https://github.com/wstein/sheetmill
  revision: 0f87dfb7b2fc5a3269e76f6c6f579cb863e430a3
  path: docs/cli.md
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
| Status | Meaning |
| :-- | :-- |
| `0` | Success |
| `1` | Invalid source or rendering diagnostic |
| `2` | Usage or file access/output failure |

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
