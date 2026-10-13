---
title: "CLI reference"
description: "Commands, input selection, exit status and every diagnostic code."
category: reference
---
| Command | Behavior |
| --- | --- |
| `smill version` | Print `smill dev` and LF |
| `smill check <file-or-directory>...` | Validate selected Markdown; no success output |
| `smill inspect <file>` | Validate one file and print canonical JSON |
| `smill import reference <file> [options]` | [Controlled conversion](imports.md), dry-run and loss reports |
| `smill serve [file-or-directory] [--open] [--port 0..65535] [--hugo /absolute/path]` | Authenticated read-only local live preview through installed Hugo 0.167.0 |
| `smill build hugo <file-or-directory>... --out <new-project> [--title <name>] [--base-path </path/>] [--edit-base-url <GitHub-edit-directory>] [--source-base-url <GitHub-blob-directory>]` | Export a self-contained [Hugo project](hugo.md); never launches Hugo |

Cobra v1.10.2 parses commands and flags. `--help` works at every command level;
top-level help (`smill --help`, `smill help`) ends with a notice pointing AI agents
to the [agent guide](agent-guide.md) and [llms.txt](../llms.txt) published with the manual;
import/build flags may precede or follow the positional input. Check/inspect
have no content options. There is no recursive or stdin mode.
Directories are scanned only one level, in sorted order, for lowercase `.md` files.
Subdirectories and other extensions are ignored; an empty batch selection is an error.
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
| `ATTR001`, `ATTR002`, `ATTR003` | Unsupported/malformed attribute; duplicate/conflict; attribute group outside a root `##`/`###` heading |
| `ID001` | Duplicate explicit/reserved ID |
| `IO001`, `IO002`, `IO003` | Input access/selection; stdout failure; output creation/write |
| `REN001`, `REN002`, `REN003` | Unsupported IR; unsafe URL; invalid/ambiguous route or base path |
| `REN101`, `REN102`, `REN103` | Inert raw HTML; images retained as unfetched links; unresolved internal link (warnings) |
| `IMP001`–`IMP104` | See [import diagnostic codes](imports.md) |

Go package tests exercise the same `Run` entrypoint as the executable, without
building a distributable CLI binary. Parser fuzz targets run via:

```sh
go test ./internal/frontmatter -run '^$' -fuzz FuzzParse -fuzztime 5s -parallel 2
go test ./internal/markdown -run '^$' -fuzz FuzzParse -fuzztime 5s -parallel 2
```

The [one-sheet E2E harness](../tests/e2e/README.md) also invokes a temporary build of
this actual entry point, then pinned Hugo and Chromium. It validates both desktop
and mobile reader behavior, no-JavaScript navigation and inert source variants.
The temporary executable is verification output, never a distribution artifact.


## Local saved-file preview

```sh
smill serve ./sheets --hugo /absolute/path/to/hugo
smill serve ./sheets/editor.md --open --port 0
```

`serve` defaults to the current directory and port 0 (available IPv4 loopback
port). It requires installed Hugo 0.167.0, resolved from the absolute `--hugo`
option or PATH; it downloads nothing. The selection is read-only and direct,
nonrecursive, with retained `os.Root` handles. Empty directories can start a
session and receive new direct `.md` files. A single-file selection never exposes
its siblings. Do not use symlinked source paths; `/private/tmp` is appropriate on
macOS. Source limits: 100 files, 8 MiB each and 32 MiB aggregate.

Stdout prints selected scope and one private bootstrap address. Open it once to
receive an HttpOnly SameSite=Strict session cookie, then use the clean root URL
in that browser. Do not share the bootstrap address. `--open` launches the fixed
platform browser opener after readiness; failure leaves the printed address and
running session available. Different ports keep distinct cookies.

Save from an external editor. Pending saves, invalid diagnostics, retained
last-valid output, removed references, publisher failures and disconnected state
are visible in trusted Studio chrome. Valid peers remain independently usable.
A new session has its own source/publishing lifecycle; same-port restart requires
the new bootstrap address. No browser editor, command runner or project-write API
exists. See [tested security boundary](studio-security.md).

Unlike batch validation's source-invalid exit 1, invalid source during `serve`
starts/keeps a diagnostic session. Usage/selection/tool/port/startup failures exit
2; normal interruption exits 0 after cleanup; runtime infrastructure failure exits
2. `STU001` diagnoses unavailable/unstable/oversized selected input in the UI;
`STU002` reports CLI session setup/runtime/cleanup failures on stderr. Compiler
and renderer diagnostic codes/spans remain unchanged. Browser state contains
source-relative names, never the host selection path or session credentials.
