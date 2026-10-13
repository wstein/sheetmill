# 006 — navi renderer

Created: 2026-10-13 · Refined: 2026-10-14 · Status: proposed contract, docs only
Depends on accepted 001–004 (schema-1 IR, diagnostics and provenance).

This revision selects explicit cards, free-input parameters and one compatibility
profile: `bash-free-input-v1`, targeting navi v2.24.0 and Bash 5.2 syntax.
It does not authorize or implement a renderer, CLI command, schema validator,
artifact generation, navi execution or deployment. No runtime compatibility test
has passed yet; the non-executing acceptance checks below remain implementation
work. See [source evidence](research.md) and the [format contract](../../docs/format/navi-profile.md).

## User scenarios and acceptance

1. After separate authorization, an author runs the proposed
   `smill build navi <file-or-directory>... --root <input-root> --out <new-dir>` and receives `.cheat`
   files containing only explicitly selected, validated commands and metadata sidecars.
2. Identical IR and configuration produce identical bytes across runs and hosts.
3. Marked cards are exported or diagnosed at their source position; command text
   is never silently rewritten. Unmarked content remains available to other renderers.
4. SheetMill, its tests and CI never execute navi, a shell or source commands.
5. Generated free-input templates carry no claim that arbitrary input is safely
   quoted or validated by navi. Using the result in navi is a separate user action.

## Profile requirements

- S1 Shell: require `cheatsheet.navi.profile: bash-free-input-v1` and exactly the
  `bash` fence language. Target Bash 5.2 grammar, checked with a non-executing
  parser after substituting inert validation sentinels for declared placeholders.
  No `sh`, `zsh`, Fish or PowerShell aliases. The profile does not configure navi:
  the consumer must configure its execution shell to compatible Bash. Do not
  inspect, invoke or modify the user's shell or navi configuration.
- S2 Selection: require an explicit source `{#id .navi}` card and matching
  `cheatsheet.navi.cards.<id>` metadata. Exactly one code block, a root-level
  fenced `bash` block in that card, is allowed; nested or additional code blocks
  are rejected. The first slice accepts one nonempty physical command line,
  with at most one terminal LF. Multiline scripts, heredocs and continuations are
  deferred, not flattened. Unmarked cards are ignored by this renderer; invalid
  marked cards fail with an error, not a successful build that silently skips them.
- S3 Payload: emit accepted code bytes unchanged, apart from adding a missing
  terminal LF as an output separator. Reject CR, NUL, controls except TAB/LF,
  Unicode category Cf, U+2028/U+2029 and U+2800 (navi's terminal field separator).
  Reject a payload starting at column one with `%`, `#`, `@`, `$`, `;` or triple
  backticks; never add indentation to evade navi parsing. Reject shell syntax
  errors, a terminal continuation and a command whose literal first word is `navi`
  (upstream rewrites that word to its own executable path). These are structural
  compatibility rules, not a dangerous-command detector.
- S4 Parameters: names match `[a-z][a-z0-9_]{0,31}` and are local to one card.
  Every navi-recognized placeholder must be a declared `<name>` used as a complete
  double-quoted shell argument, for example `"<sheet>"`. Reject escaped tokens,
  interpolation inside a larger word, comments, substitutions or other quoting
  contexts. Every declaration must be used; repeated use of a name is allowed.
  There is no literal-placeholder escape in v1. Detect navi's broader token syntax
  before checking this narrower profile; reject unsupported names rather than
  treating them as literal text. Ordinary shell redirections such as `cat < file`
  are syntax, not unbalanced placeholders, and remain eligible.
- S5 Metadata: follow the closed proposed schema in the format contract. Parameters
  have `type: string|path` and a nonempty `description`; both are authoring hints
  preserved in the sidecar, not runtime validation. Zero-parameter cards use
  `params: {}`. No defaults, enums, suggestion commands, maps, previews,
  multi-selection, dependencies or `@` inheritance. Requested deferred features
  produce `NAV005`; unknown keys produce `NAV006`. No `$` provider lines are emitted.
- S6 Never execute: future renderer tests compare independently authored golden
  bytes, validate Bash syntax and model the pinned navi parser/interpolation as
  pure data transformations. No navi CLI, shell, finder or snippet execution,
  including print/preview modes. An import/dependency audit must cover indirect
  execution paths, not just direct `os/exec`, syscall-exec or plugin imports.
- S7 Preservation: consume schema-1 IR without reparsing Markdown or changing its
  canonical representation. Preserve complete IR metadata and the selected card
  IDs, descriptions, parameters and profile in a deterministic versioned
  `<sheet>.navi.json` sidecar. Unknown non-navi metadata stays there; unknown navi
  keys are errors. Prose/layout are intentionally not executable entries.
- S8 Identity and provenance: emit `;` metacomments for profile, target navi version,
  title, relative source path, SheetMill version, attribution and license. Encode
  each header value as one JSON string on one physical line. Never use `#` for
  provenance. A missing license/attribution is explicitly `not stated` and warns
  `NAV102`. Use one context per card as specified in the format contract; no shared
  suggestion scope. Terminate encoded contexts with `!` and validate the pinned
  navi item hash across the complete export batch (`NAV008` on duplicates).
  Source titles/paths cannot inject commands or finder fields.

## Runtime boundary

Navi v2.24.0 replaces placeholders with raw text, including environment-provided
values. `"<sheet>"` preserves ordinary spaces but does not make values containing
quotes, substitutions or shell operators safe. SheetMill cannot validate future
free input and must not promise argument-boundary preservation for arbitrary
values. Generated provenance and user documentation must say that values are
trusted shell text, not sandboxed data. Leading `-` can become a program option;
path authors should use an application-supported `--` or `./` prefix where useful.
There is no universal option-injection fix and no runtime-value warning check.

## Acceptance matrix (required later; not implemented or executed here)

| Case | Expected contract |
| --- | --- |
| Explicit card, one Bash command, declared quoted parameter | One entry, zero providers; exact payload and sidecar metadata |
| Same source twice, input order reversed, different host paths | Same relative identities, output ordering and bytes |
| Unmarked Bash example | No executable entry; source remains unchanged |
| Missing explicit ID, metadata or block; nested/multiple blocks; wrong language | `NAV007`, fail before writing |
| Metadata refers to missing or unmarked card | `NAV004`, fail before writing |
| `<Name>`, `<a-b>`, escaped token, embedded `prefix<name>` | `NAV003`; never silently literalized |
| Undeclared `<name>` or unused declaration | `NAV004` |
| `cat < input.txt`, `echo ok > output.txt` | Accepted redirections with empty params |
| Unterminated quotes, incomplete pipeline, continuation, multiline/heredoc | `NAV002` |
| Payload starts `%`, `#`, `@`, `$`, `;` or triple backticks | `NAV002`; fences do not repair it |
| CR, NUL, Cf, line/paragraph separators, navi terminal separator | `NAV002` in payload; `NAV006` in context/display metadata |
| Defaults, enums, providers, map/preview, inheritance, multi-selection | `NAV005`; no runtime providers |
| Unknown profile/key/type; invalid description | `NAV001` / `NAV006` / `NAV004` |
| Duplicate card titles, duplicate basenames in separate directories | Distinct contexts; pinned parser retains every entry |
| IDs `a`/`ab`, titles `bC`/`C`, same sheet and no tags | Terminated contexts avoid identical prehash text; both entries retained |
| Duplicate pinned navi item hashes in a batch | `NAV008`, fail before writing, even when context strings differ |
| Two cards use `<path>` | Separate parameter declarations; no shared suggestions |
| Unicode titles and values; spaces, quotes, `$()`, semicolons, leading `-` | Pure substitution fixtures record raw-text behavior; never execute results or claim safety |
| Literal command first word `navi` | `NAV002`; no executable-path rewrite |
| Missing license/attribution; multiline provenance | `NAV102`; JSON-quoted single-line `;` headers, metadata preserved |
| Marked invalid card alongside valid card | Errors prevent all output; no partially successful export |

## Requirements and diagnostics

- FR-001: Consume canonical IR directly, with no source evaluation or fetching.
- FR-002: Record profile/target and output sidecar schema version `1` in outputs.
- FR-003: Require a new output directory; refuse existing/symlink destinations;
  validate every input and destination collision before writing, no overwrite.
  Require an explicit `--root`; resolve inputs within it, rejecting traversal,
  outside-root inputs and symlinks. The root itself never enters output bytes.
- FR-004: Order sheets by normalized relative path and cards by source order.
  Preserve relative directory structure for `.cheat` and sidecar filenames. Reject
  duplicate relative identities/output paths. No clocks or absolute host paths.
- FR-005: Preserve metadata and provenance per S7–S8.
- FR-006: Document proposed codes when implementation is authorized:
  `NAV001` unsupported profile; `NAV002` incompatible command text;
  `NAV003` unsupported placeholder syntax/context; `NAV004` invalid parameter or
  card reference; `NAV005` deferred capability; `NAV006` invalid/unknown metadata;
  `NAV007` invalid marked-card shape/selection; `NAV008` navi identity collision;
  `NAV102` missing provenance warning.
  Former draft `NAV101` skip and `NAV103` runtime-value warning are withdrawn.
  Shared `REN001`/`IO003` remain applicable. Any error prevents writing.
- FR-007: Obtain separate implementation, generation and publication authorization.
  This documentation change grants none and adds no placeholder CLI command.

## Success criteria and remaining work

Before claiming this profile tested, implement every acceptance row, pure parsing
of the generated output and pure substitution fixtures grounded in the pinned
upstream source. Assert entry count/context/title/payload, zero providers, complete
sidecar metadata, deterministic hashes and execution-path audit. Fuzz payloads,
headers and metadata for structural injection. Record the exact tested revision
and independently review it. A handwritten parser plus golden snapshots alone
is not evidence of compatibility with an unexamined upstream release.

The profile/schema decisions above are proposed for docs review. Runtime tests,
Bash parser dependency selection/pinning and all renderer work remain pending.
Multiline scripts, literal placeholder escaping, other shells, enum/default UI,
providers and round-trip navi import require later specifications.
