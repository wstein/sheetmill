# 006 — navi renderer

Created: 2026-10-13 · Refined: 2026-10-14 · Status: implemented, awaiting independent review
Depends on accepted 001–004 (schema-1 IR, diagnostics and provenance).

Explicit cards, free-input parameters and per-shell profiles, targeting navi
v2.24.0. The normative user contract is the
[format page](../../docs/format/navi-profile.md); this file records decisions and
requirements. Source evidence is in [research.md](research.md). Implementation was
authorized by the user on 2026-10-14. Navi execution, publication of generated
cheats and distribution remain gated, and no runtime compatibility test with real
navi has been run.

## Decisions (user, 2026-10-14)

- Output is `.cheat`, one file per sheet.
- A card is exported only with `{.navi}` and an explicit `{#id}`; the IR records
  this as `id_explicit` (omitted when false).
- Profiles: `bash` first (default), then `zsh`, `fish` and optional `nu`. navi uses
  one configured shell and the fence label does not select it, so a build targets
  one profile and records it. `bash-free-input-v1` was the contract proposal's name
  for the bash profile; its Bash 5.2 grammar parser requirement is replaced by a
  heuristic quote check until a pure Go parser is pinned.
- A marked card that cannot be exported is a **warning**: it is skipped and never
  rewritten. Metadata, profile and batch-identity problems are errors that write
  nothing.
- navi, a shell and exported commands are never run, including print or preview
  modes.

## Requirements

- S1 Shell: profiles above. Fence languages accepted per profile are listed in the
  format page; any other language skips the card (`NAV101`).
- S2 Selection: `{.navi}` plus explicit `{#id}`, exactly one fenced code block.
  Multi-line commands are allowed; blank lines are not (navi turns them into
  separators).
- S3 Payload: exported unchanged. Skip the card for CR, NUL and other controls,
  Cf characters, U+2028/U+2029, U+2800, a column-0 `%`, `@`, `;`, `#`, `$` or fence
  line, a trailing continuation backslash, an unterminated quote and a leading
  `navi` word (`NAV102`). Nothing is indented or hidden to evade navi's parser.
- S4 Parameters: `<name>` with `[a-z][a-z0-9_]{0,31}`; declared per card under
  `cheatsheet.navi.cards.<id>.params`; every declaration used. Look-alike
  placeholders skip the card (`NAV103`, `NAV104`). Redirections are shell, not
  placeholders. There is no literal-placeholder escape in this version.
- S5 Metadata: closed schema, `type` in `string|int|path`, optional single-line
  `description`. Unknown keys, dangling card IDs and unsupported types are `NAV006`;
  inconsistent declarations of a name in one sheet are `NAV007`. Defaults, choice
  lists, providers, previews, multi-selection and shared contexts are out of scope;
  no `$` lines are emitted.
- S6 Never execute: tests compare bytes and use a pure Go model of navi's line
  grammar and entry deduplication; a lint test forbids `os/exec`, `syscall`,
  `plugin` and network imports in the package.
- S7 Preservation: consume schema-1 IR without reparsing Markdown. Metadata outside
  the navi namespace is **not yet preserved**: the planned versioned
  `<sheet>.navi.json` sidecar is a remaining task.
- S8 Identity and provenance: each card gets a unique first tag
  `sheetmill:<path>#<id>!` (percent-encoded, terminated), followed by `tool` and
  `tags`. The batch is checked for equal pinned item hashes (`NAV008`) because navi
  silently drops later entries with the same hash. Header comments carry source,
  title, shell, license (or "not stated", `NAV105`) and attribution as single
  lines.

## Runtime boundary

navi replaces placeholders with raw text, including environment values. Quoted
placeholders keep ordinary spaces together but do not make arbitrary values safe;
types are not enforced. Generated headers and the format page state this.

## Acceptance (implemented in tests)

| Case | Contract |
| --- | --- |
| Explicit card, declared parameter | One entry with a unique context, exact payload |
| Same title in one sheet; same tags and titles across sheets | Both kept; none dropped by the navi model |
| Equal path and ID in two sheets | `NAV008`, nothing written |
| Missing explicit ID, wrong block count, other language | `NAV101` warning, card skipped |
| Unterminated quote, markers, controls, U+2800, `navi` word, blank line, trailing `\` | `NAV102` warning |
| `<Name>`, `<a-b>`, undeclared, unused | `NAV103` / `NAV104` warning |
| Redirections, heredocs, comments with quotes, leading dash, Unicode | Exported unchanged |
| Unknown metadata, dangling card, bad type, wrong `shell`, bad tag | `NAV006` error |
| Conflicting declarations; unsupported profile | `NAV007`; `NAV001` |
| Header value with newline or `%` | Collapsed to one line, no new entry |
| Input order and host path | Same bytes; no absolute path in output |
| Sheet or selection with nothing exportable | `NAV106` warning; build fails if no file results |

## Diagnostics

`NAV001`, `NAV006`, `NAV007`, `NAV008` errors; `NAV101`–`NAV106` warnings; shared
`REN001`, `IO001` and `IO003` apply. Meanings are in the format page.

## Remaining work

Independent review before merge; metadata sidecar; a pure Go Bash/zsh/fish parser
to replace the heuristic quote check (dependency to be justified and pinned);
choice lists and shared contexts, which need their own design; nushell raw-string
analysis. A non-executing parse check against real navi, publication of generated
cheats and distribution builds are separately gated.
