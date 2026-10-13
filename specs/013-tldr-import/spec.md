# 013 — Controlled TLDR import

Created: 2026-10-13 · Status: draft, specification only · Depends on: 002, 003

Implementation requires separate authorization. Statements about the tldr-pages
format and licenses are from author knowledge, to verify against a pinned upstream
revision before implementation.

## User scenarios and acceptance

1. P1: A user runs `smill import tldr pages/common/tar.md` with the same required
   provenance flags as Reference import (`--repository`, `--source-path`,
   `--revision`, `--license`, `--attribution`) and gets one canonical GFM+ sheet.
2. P1: The page title becomes the sheet title, the `>` lines become the sheet
   description and links, and each example becomes a card. `{{placeholder}}`
   tokens remain visible, literal content, never expanded, validated or executed.
3. P1: Platform and language come from the source path (`pages/common/`,
   `pages.de/linux/`) and are recorded as metadata, not guessed from content.
4. P1: A user previews with `--dry-run --report json`. Unknown or unsupported
   constructs appear in the loss report with source positions; none is dropped.
5. P2: Rerunning conversion with identical source and provenance is byte-identical.
   Existing outputs are refused; there is no force mode.

## Page structure mapping (to verify)

| Source construct | SheetMill result |
| --- | --- |
| `# name` (one H1) | Sheet title; page name also kept as `source.original_metadata` |
| `> description` lines | Sheet description paragraph(s) |
| `> More information: <url>.` | Link retained as text and as inert metadata; never fetched |
| `- description:` line | Card heading (level fixed by core format) |
| following backticked command line | Fenced code block in that card, content verbatim |
| Platform/language path segments | `source.original_metadata` keys `platform`, `language` |

All examples sit under one section named `Examples` (see Open decisions).

## Placeholder representation

`{{path/to/file}}` is kept byte-for-byte inside fenced code, where core GFM+ does
not interpret attributes. Fences are always used for commands, so placeholders
never reach prose attribute parsing (`{.x #y}`). Placeholders found in prose (for
example in a description, outside backticks) are wrapped in an inline code span and
reported as `TLD104`. A literal `{.x}` or `{#y}` in prose is escaped or code-spanned
and reported, so imported text can never become a SheetMill attribute. Optional
placeholder markers (`{{[-f|--file]}}`) and escaped braces are kept verbatim.
The report lists each distinct placeholder name and count. No substitution,
shell parsing or command validation occurs.

## Requirements

- FR-001: Only explicit `smill import tldr <file>`; one local file; no
  auto-detection, directory, glob, recursive or batch input.
- FR-002: Same explicit, validated provenance flags and rules as FR-002 of 003;
  `--out`/`--dry-run`/`--report json` and exclusive creation behave as in 003.
- FR-003: Deterministic mapping above; no clocks, locale or map-order dependence.
- FR-004: Platform and language derive only from the clean `--source-path`.
  Unrecognized path layouts are reported (`TLD102`), not guessed.
- FR-005: Preserve `{{placeholder}}` text and command bytes verbatim; no expansion,
  execution, shell tokenizing or suggestion generation.
- FR-006: Unknown, malformed or unsupported constructs (extra headings, tables,
  HTML, unmatched example lines, multiple backtick spans, trailing text) are kept
  as inert content or metadata and reported with source-positioned codes. Content
  that cannot be represented safely fails with an error and writes nothing.
- FR-007: Validate converted bytes with the canonical parser before any write.
- FR-008: No network access and no link fetching. The "More information" URL is
  data only.
- FR-009: Retain license and attribution as source metadata; conversion never
  grants redistribution rights. Content license (CC BY 4.0, to verify) is checked
  separately from the code license (MIT, to verify).
- FR-010: Diagnostics use the new `TLD` range, documented in `docs/imports.md`
  before release; core `IMP001`–`IMP003` are reused for shared provenance failures.
- FR-011: Tests use independently authored synthetic fixtures and byte-for-byte
  golden Markdown and JSON reports; no upstream page enters the repository
  without license review. A fuzz target covers the converter.

## Diagnostic codes (proposed)

`IMP` already uses 001–003 and 101–104, so TLDR gets its own `TLD` prefix.

| Code | Meaning |
| --- | --- |
| `TLD001` | Page has no single H1 title (error) |
| `TLD002` | Example description without a command (error) |
| `TLD101` | Unknown line or block type retained inert (warning) |
| `TLD102` | Source path layout not recognized; no platform/language (warning) |
| `TLD103` | Command line has several or unbalanced backtick spans (warning) |
| `TLD104` | Placeholder or brace text in prose code-spanned (warning) |
| `TLD105` | Raw HTML retained as inert data (warning) |
| `TLD106` | "More information" link recorded, not fetched (warning) |

## Success criteria

Converted output reparses; identical inputs give identical bytes and reports.
Every placeholder in a fixture appears unchanged in the output. Each injected
unsupported construct yields a finding or an error. Dry-run writes nothing. The
fuzz target finds no panic or non-reparseable successful output.

## Open decisions

- One card per example, or grouping consecutive examples by shared theme.
- Mapping of long multi-line descriptions: sheet description versus a prose block.
- Merging multiple platform variants (`common` plus `linux`) into one sheet; the
  default is no merge, one file per sheet.
- Whether `Examples` stays a fixed section name or follows page language.

Out of scope: fetching, batch/recursive import, Reference/navi/Devhints adapters,
rendering, publication and legal decisions on republishing upstream content.
