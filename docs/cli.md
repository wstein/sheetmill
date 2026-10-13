# CLI

| Command | Behavior |
| --- | --- |
| `smill version` | Print `smill dev` and LF |
| `smill check <file-or-directory>...` | Validate selected Markdown; no success output |
| `smill inspect <file>` | Validate one file and print canonical JSON |
| `smill import reference <file> [options]` | [Controlled conversion](imports.md), dry-run and loss reports |
| `smill build hugo <file-or-directory>... --out <new-project> [--title <name>]` | Export a self-contained [Hugo project](hugo.md); never launches Hugo |

Cobra v1.10.2 parses commands and flags. `--help` works at every command level;
import/build flags may precede or follow the positional input. Check/inspect
have no content options. There is no recursive or stdin mode.
Directories are scanned only one level, in sorted order, for lowercase `.md` files.
Subdirectories and other extensions are ignored; an empty selection is an error.
Duplicate explicitly selected paths are checked once. Selected files must be
regular files. Symlinks in selected paths or `.md` directory entries are rejected,
including ancestor symlinks. On macOS use `/private/tmp` instead of `/tmp` for
scratch files, or resolve the explicit path yourself first. Input reads are confined
with `os.Root` and limited to 8 MiB plus one byte for oversized-input detection.
No implicit filesystem crawling occurs. Source symlinks are unsupported; this
is not a sandbox against an adversary concurrently mutating all parent directories.

Stdout contains requested version/JSON/help data and a build export summary. Diagnostics use stderr:
`path:line:column: CODE: actionable message`. Validations across explicit inputs
continue after errors. Exit status precedence is IO/usage (2), then invalid source
(1), then success (0). Broken stdout also produces exit 2.

| Code | Meaning |
| --- | --- |
| `SRC001`, `SRC002` | Invalid UTF-8; source size limit |
| `FM001`, `FM002` | Missing opening/closing frontmatter delimiter |
| `FM003` | Invalid YAML document/root |
| `FM004` | Duplicate mapping key |
| `FM005` | Unsupported YAML construct, type or nesting |
| `VAL001` | Missing/invalid title |
| `VAL002` | Invalid general known metadata field |
| `VAL003` | Invalid cheatsheet metadata/schema |
| `MD001`, `MD002` | Card without section; empty structural heading |
| `ATTR001`, `ATTR002` | Unsupported/malformed attribute; duplicate/conflict |
| `ID001` | Duplicate explicit/reserved ID |
| `IO001`, `IO002`, `IO003` | Input access/selection; stdout failure; output creation/write |
| `REN001`, `REN002`, `REN003` | Unsupported IR; unsafe URL; invalid/ambiguous route |
| `REN101`, `REN102`, `REN103` | Inert raw HTML; images retained as unfetched links; unresolved internal link (warnings) |
| `IMP001`–`IMP104` | See [import diagnostic codes](imports.md) |

Go package tests exercise the same `Run` entrypoint as the executable, without
building a distributable CLI binary. Parser fuzz targets run via:

```sh
go test ./internal/frontmatter -run '^$' -fuzz FuzzParse -fuzztime 5s -parallel 2
go test ./internal/markdown -run '^$' -fuzz FuzzParse -fuzztime 5s -parallel 2
```
