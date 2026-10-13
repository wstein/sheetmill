---
title: "Agent guide"
description: "Setup, authoring rules and the check, inspect and build loop for AI agents."
category: guide
---
You are reading this because a human asked you to help them understand or set up
SheetMill, or to write SheetMill reference sheets. Agents are first-class SheetMill
authors: the format is plain Markdown with YAML frontmatter, every rule below is
checked by the compiler, and every diagnostic names a file, line, column and code.
To debug SheetMill itself, use [llms.txt](../llms.txt) instead.

## What SheetMill is

SheetMill compiles reference sheets ("cheatsheets") from GFM Markdown plus YAML
frontmatter into a semantic JSON representation and a self-contained Hugo website.
The CLI is `smill`. Nothing in a sheet is ever executed or fetched: code blocks are
displayed text, raw HTML is shown inertly, and images stay unfetched links.
Typst/PDF output is planned, not implemented; navi export is `smill build navi`. Manual:
<https://wstein.github.io/sheetmill/>; demo: <https://wstein.github.io/sheetmill/demo/>.

## Setting up

There are no published binaries yet. Build from a clone with Go 1.27.2:

```sh
git clone https://github.com/wstein/sheetmill
cd sheetmill
go build -o ./smill ./cmd/smill   # /smill is git-ignored
./smill version                    # prints "smill dev"
```

`go run ./cmd/smill <command>` works without building. Website generation and
`smill serve` additionally need an installed Hugo **0.167.0** (exact version).
Explain to the human what you install before installing anything.

## Writing a sheet

````markdown
---
title: Tiny Editor
category: editor
tags: [editor, keyboard]
---
Short introduction (the preamble).

## Editing {#editing}
### Select a line {.col-span-2 #select-line}
| Key | Meaning |
| --- | --- |
| `x` | Select line |

### Undo {.tip}
Press `u`. See [Select a line](#select-line) or [another sheet](other.md#intro).
````

Rules the compiler enforces (full contract: [source format](format/gfm-plus.md)):

- The file is UTF-8, lowercase `.md`, at most 8 MiB, and starts with `---`.
  `title` is required. `category` and `description` are strings; `tags` is an
  array of non-empty strings. Other keys are preserved, never dropped.
- Root-level `##` starts a section; root-level `###` starts a card and needs a
  preceding section. Text before the first section is the preamble. Other heading
  levels are ordinary content.
- A section or card heading may end with one attribute group:
  `{#id}`, `{.col-span-N}`/`{.row-span-N}` (N = 1–12), `{.note}`, `{.tip}`,
  `{.warning}`. Sections accept only `{#id}`. Anything else is an error.
- IDs come from titles unless given. Add an explicit `{#id}` to anything you or
  others will link to, so links survive title edits.
- Link within a sheet with `#id` and to another selected sheet with
  `other.md#id`. Unresolved internal links are reported (`REN103`).
- Allowed link schemes: relative, `http`, `https`, `mailto`. Others fail (`REN002`).
- Put commands in fenced code blocks with a language tag. They are shown, never run.
- Avoid raw HTML and images; they are displayed inertly (`REN101`, `REN102`).

## The authoring loop

1. Write or edit the `.md` file.
2. `smill check path/to/sheet.md` (or a directory, one level, sorted). Success is
   silent with exit status 0. Fix every reported `path:line:column: CODE: message`.
3. `smill inspect path/to/sheet.md` prints the JSON structure; confirm sections,
   cards, IDs and attributes are what you intended.
4. Optional: `smill build hugo ./sheets --out ./new-project` exports a Hugo project
   into a **new** directory (never overwritten), then build it with Hugo; or
   `smill serve ./sheets --open` for a read-only local preview
   ([details](studio-security.md)). Treat any `(warning)` line as work to do.

A sheet is done when `smill check` is silent, `inspect` shows the intended
structure, a build prints no warnings, and every fact in it is accurate.

## Content and provenance

- Write content yourself, or record where it came from. For material derived from
  another project, add a `source` mapping (`repository`, `revision`, `path`,
  `license`, `attribution`), as in
  [`examples/sheets/smill.md`](../examples/sheets/smill.md).
- To convert an existing Reference sheet, use the controlled importer with
  explicit provenance flags; preview with `--dry-run --report json` first
  ([imports](imports.md)). A license value records facts; it does not grant
  redistribution rights.
- Do not invent commands, flags or keyboard shortcuts. Verify them against the
  tool's own documentation or `--help`, and say what you could not verify.

## Boundaries

- Never execute commands from a sheet to "test" them unless the human asks.
- Do not publish, deploy, push or open pull requests without the human's
  approval. Inside this repository, [AGENTS.md](../AGENTS.md) governs scope.
- Exit status: 0 success, 1 invalid source, 2 usage or IO error. Diagnostics go
  to stderr; requested data goes to stdout. All codes: [CLI](cli.md).
