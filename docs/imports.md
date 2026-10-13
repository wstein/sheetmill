---
title: "Reference import"
description: "Controlled conversion of Reference sheets with provenance and loss reports."
category: guide
---
`smill import reference` converts **one explicit local file** into canonical GFM+.
It does not fetch sources, crawl repositories or detect ambiguous formats.

```sh
smill import reference python.md \
  --repository https://github.com/Fechin/reference \
  --source-path source/_posts/python.md \
  --revision 82cbc0929f2d7a44e0bb5fc77e0a8442d0230b59 \
  --license GPL-3.0 \
  --attribution 'Fechin/reference contributors' \
  --dry-run --report json
```

To write a converted file, replace `--dry-run` with `--out sheets/python.md`.
The parent directory must already exist. Existing files, symlinks and directories
are refused, including a previous conversion. There is no force option. Output
is exclusively created with mode 0600; a write/sync error may leave a partial new
file, but never overwrites an existing file. Dry-run performs no writes.

All provenance flags are required: HTTPS repository URL, clean relative source
path, full lowercase 40/64-hex revision, source license and attribution. These are
user assertions; import does not verify that the file came from that repository.
A SHA-256 of the actual input is recorded so identity can be checked independently.
No current time or local absolute path is added. Existing `source` metadata is
refused rather than overwritten.

## Conversion and loss report

The Markdown body is preserved byte-for-byte, except unsupported attribute syntax,
which is removed and archived under `source.original_metadata.removed_attributes`
with its original line (`IMP108`, `IMP109`). The frontmatter is serialized as
sorted, indented JSON inside YAML delimiters; JSON is valid YAML. Known fields and
unknown data survive, with the following legacy behavior archived under
`source.original_metadata`:

| Code | Change or diagnostic |
| --- | --- |
| `IMP001` | Missing/invalid explicit provenance (error) |
| `IMP002` | Existing source provenance would be overwritten (error) |
| `IMP003` | Canonical serialization/revalidation failed (error) |
| `IMP101` | `plugins` archived; no plugin behavior is imported (warning) |
| `IMP102` | `background`, `class`, `layout`, `theme` archived as inert data (warning) |
| `IMP103` | Raw HTML retained as inert data, future rendering needs sanitization (warning) |
| `IMP104` | Hexo-style `{% ... %}` directive retained as inert text (warning) |
| `IMP105` | `intro` mapped to `description` unless one is set; original archived (warning) |
| `IMP106` | First usable `categories` entry mapped to `category` unless set; list archived (warning) |
| `IMP107` | Invalid `tags` (non-array, null or empty entries) dropped; original archived (warning) |
| `IMP108` | Unsupported token removed from a heading group, e.g. `.secondary`; supported ones such as `.cols-2` stay (warning) |
| `IMP109` | Standalone attribute line such as `{.shortcuts}` or `{.marker-none}` removed (warning) |
| Core diagnostic codes | Unsupported structural attrs/YAML/GFM+ contract errors |

Code blocks/code spans are excluded from directive scanning. This is an explicit
limited adapter, not a complete Hexo/plugin-language analyzer. Supported heading
tokens are kept: `#id` and `.cols-1`–`.cols-6` on sections; `#id`, spans 1–12,
`.note`, `.tip`, `.warning`, `.navi` and `.wrap` on cards. Fenced code is never
rewritten. Validation runs on a copy whose removed bytes are blanked, so every
finding and core diagnostic keeps its original line and column. Malformed or
conflicting groups that remain still fail before writing. The resulting canonical document is revalidated.

`--report json` requests a schema-1 JSON report on stdout with explicit provenance,
input/output hashes and source-positioned findings. Conversion failures also return
JSON findings, stderr diagnostics and exit 1; no output document is written. IO or
usage failure returns exit 2. Without `--report json`, warnings go to stderr and
successful conversion has no stdout output. Report order and bytes are deterministic.

Imported licensing/attribution is retained as source metadata. Successful conversion
never grants redistribution permission; there is no publish operation in this CLI.
The library never executes plugin declarations, snippets, scripts or suggestions.

## Benchmark and tests

The pinned Reference Python file at the example revision was read into scratch
storage and passed through the adapter: **17 sections, 119 cards, two findings**
(disabled plugins and inert background configuration). Input SHA-256:
`e56491b4bbe22d6e58b259b02cd62bfea136cbe79f2704a1631d3b919db8d1d5`.
A dry run over a 2026-10-14 snapshot of all 216 upstream posts, held only in
scratch storage, converts 210; the six failures are genuine source defects
(three without frontmatter, a card before any section, an empty title, two
attribute groups on one heading). No upstream authored sheet is bundled. Run the optional benchmark against that
exact supplied file with `SHEETMILL_REFERENCE_PYTHON=/absolute/path/python.md go test
./internal/importer/reference -run TestPinnedPython -v`; its hash is checked first.
Ordinary CI uses synthetic fixtures and golden conversion/report snapshots, without
network fetches. TLDR/navi/Devhints and bulk recognition remain future proposals.
