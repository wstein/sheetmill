# 006 — navi renderer

Created: 2026-10-13 · Status: draft, specification only · Depends on: 002-gfm-parser
and accepted 001–004 (IR schema 1, diagnostics, provenance)

Statements marked "to verify" rest on general knowledge of the navi cheat format and
must be checked against the pinned navi version before the contract is accepted.
Exporting is text generation only: navi is never run.

## User scenarios and acceptance

1. P1: An author runs `smill build navi <file-or-directory>... --out <new-dir>` and gets
   `.cheat` files containing only explicitly marked, validated commands.
2. P1: Given identical IR and configuration, the output bytes are identical across
   runs, input orders and hosts.
3. P1: Anything not exportable is diagnosed with a source position; nothing is
   dropped silently and nothing is rewritten to "make it safe".
4. P1: No exported command is ever executed by SheetMill, its tests or CI.

## Safety requirements (to accept before authorization; S1, S4 and S5 hold open decisions)

- S1 Shells: open decision. Candidate: POSIX `sh` syntax only (a conservative subset,
  as accepted by navi's default shell, to verify); `bash`/`zsh` as later, separately
  specified profiles. Unlisted language tags are not exportable (`NAV101`).
- S2 Exportable cards: only cards carrying the `{.navi}` class AND containing exactly
  one fenced code block whose language tag is in the supported shell list. A card
  without `{.navi}`, without a code block, with several code blocks, or with an
  unsupported tag is skipped with `NAV101` (warning) if marked, silently ignored
  only if unmarked. Never infer exportability from content.
- S3 Command text: each line is exported byte-for-byte; no variable expansion, no
  rewriting. The command is rejected (`NAV002`, error) if it contains NUL or other
  control characters except TAB and LF, a carriage return, an unterminated quote,
  an unbalanced `<`/`>` placeholder, a line ending in an unescaped backslash
  at EOF, or navi syntax that would change meaning (a line starting with `%`, `#`,
  `$` or `;` outside the permitted header position, to verify).
  SheetMill does not claim to detect "dangerous" commands; the card's author owns
  that. It guarantees that export structure (headers, parameters) cannot be altered
  by command text.
- S4 Parameters: a placeholder is exactly `<name>` with `name` matching
  `[a-z][a-z0-9_]{0,31}`. Text that merely resembles one (`<Name>`, `< a>`, `<a-b>`,
  `<>`) is a `NAV003` error rather than literal text, unless escaped as the source
  specifies (rule to define in T003). Every placeholder must be declared in the
  card's typed metadata and every declaration used (`NAV004`).
- S5 Typed parameters map from metadata, not from prose:
  `cheatsheet.navi.params.<name>.type` in `string|int|path|enum`, with optional
  `default` (scalar), `values` (non-empty array for `enum`) and `description`
  (string). Schema is an open decision to finalize in T003, including how card-level
  metadata is attached in IR (to verify against docs/format; today `cheatsheet` is
  sheet-level and `.navi` is a reserved class only). Enum values and defaults
  must be single-line, free of control characters and, when `values` become
  suggestion lists, are emitted as quoted literals, never as shell code. Dynamic
  suggestions (navi `$ name: command` lines) are not generated (`NAV005`
  if metadata asks for them).
- S6 Never execute: the renderer, its unit/golden tests, fuzz targets and CI never
  invoke navi, a shell or any exported command. Tests compare bytes and use a pure Go
  parser of the generated `.cheat` grammar. An import-lint test fails if
  `os/exec`, `syscall` exec or `plugin` is referenced under the navi package.
- S7 Unknown metadata: unknown keys under `cheatsheet.navi` and unknown params
  fields are diagnosed (`NAV006`, error) or preserved in a sidecar
  `<sheet>.navi.json`; never dropped. Metadata outside `navi` is preserved in the
  same sidecar.
- S8 Provenance: each exported file begins with comment-header lines carrying
  sheet title, source path, tool name/version of SheetMill, license and attribution
  from IR/provenance (comment syntax to verify). Header values are escaped to a single
  line; a missing license is stated as "not stated" with warning `NAV102`.

## Negative scenarios (table-driven tests; each also run for every supported shell)

| Case | Input fragment | Expected |
| --- | --- | --- |
| Single/double quote | `echo 'a` / `echo "a` | `NAV002` unterminated quote |
| Command substitution | `echo $(id)` in a parameter default | `NAV004`/`NAV005`: values are never shell code; command-body use is verbatim, with no expansion by SheetMill |
| Backticks | `` echo `id` `` as parameter value | rejected as above; verbatim in body |
| Semicolon, `&&`, `\|\|` | `a; b`, `a && b` in a default or enum value | rejected (`NAV004`); allowed verbatim in body |
| Newline | default `"a\nb"`, CR, U+2028 | `NAV004` (single-line only) |
| Unicode | bidi controls (U+202E), zero-width, NUL | `NAV002`; ordinary non-ASCII letters accepted |
| Leading dash | value `-rf`, `--x` for `path`/`string` | warned `NAV103` unless the placeholder follows `--` in the body (rule T003) |
| Empty params | `<>`, `<a>` with empty default, empty enum | `NAV003` / `NAV004` |
| Undeclared / unused | `<x>` without metadata; metadata without `<x>` | `NAV004` |
| Header injection | title with newline or `%` start | single-line escaped; never creates a header/command line |
| Navi marker in body | line `% tags`, `$ v: cmd` inside code | `NAV002` |
| Unknown language | `` ```python `` with `{.navi}` | `NAV101` skip |
| Unknown metadata | `navi.params.a.foo: 1` | `NAV006` |

## Requirements

- FR-001: Consume schema-1 IR directly; never parse source formats or run snippets.
- FR-002: Escaping, layout and output-version contract is S1–S8; the output format
  version is recorded in the header and in docs before implementation.
- FR-003: Require a new output directory; refuse existing/symlink destinations,
  validate all inputs before writing, no overwrite, no implicit downloads.
- FR-004: Deterministic bytes: stable ordering by sheet path then card order; no
  clocks, host paths or environment data; pin navi (`.navi-version`) only for an
  optional, separately approved parse-compatibility check, never to run commands
  (to verify whether such a check is possible without execution; otherwise rely on
  the pure parser).
- FR-005: Retain source attribution and license metadata (S8).
- FR-006: Diagnostics proposed, documented in `docs/cli.md` and `docs/navi.md`:
  - `NAV001` (error): unsupported shell profile or configuration.
  - `NAV002` (error): unsafe or unterminated command text.
  - `NAV003` (error): malformed placeholder.
  - `NAV004` (error): undeclared/unused or invalid parameter value or type.
  - `NAV005` (error): dynamic suggestion requested.
  - `NAV006` (error): unknown navi metadata.
  - `NAV101` (warning): marked card skipped (unsupported tag, shape).
  - `NAV102` (warning): license or provenance not stated.
  - `NAV103` (warning): value begins with `-` and may be parsed as an option.
  - Shared `REN001`, `IO003` reused. Any error fails the build before writing.
- FR-007: Obtain separate implementation, generation and publication authorization;
  this specification grants none.

## Success criteria

After authorization: every negative row passes as a table-driven test; golden
`.cheat` bytes are independently authored; a fuzz target over command text,
parameter values and headers asserts that the pure parser always recovers exactly
the intended structure (one header block, commands and parameters unchanged);
output hashes match across runs; an independent reviewer approves the final
revision, including the no-execution lint. No build or publication happens now.

## Open decisions

Supported shell profile(s) (S1); card-level metadata location and typed parameter
schema (S5); placeholder escape rule and leading-dash policy (S4, T003); whether a
non-executing navi parse check exists; navi version to pin.
