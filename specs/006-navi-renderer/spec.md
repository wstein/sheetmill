# 006 — navi renderer

Created: 2026-10-13 · Refined: 2026-10-14 · Status: exporter/review fixes merged; T009 sidecars implemented; independent review pending
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
  for the bash profile. Bash/Zsh quote diagnostics now use only the pinned
  `mvdan.cc/sh/v3/syntax` v3.14.1 parser, without executing code. Full syntax
  validation and a suitable Fish parser remain pending.
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
  `cheatsheet.navi.cards.<id>.params`; every declaration used. Placeholders are
  recognized with navi's own broader regex first (`\\?<(\w[\w\d\-_]*)>`, Unicode
  `\w`, so `<1>`, `<ünï>` and `\<name>` count), then rejected when outside the
  narrower grammar (`NAV103`, `NAV104`). Redirections are shell, not
  placeholders. There is no literal-placeholder escape in this version.
- S5 Metadata: closed schema, `type` in `string|int|path`, optional single-line
  `description`. Unknown keys, dangling card IDs and unsupported types are `NAV006`.
  Parameters are card-local (`NAV007`, the sheet-wide conflict, is retired because
  each card has its own context). Defaults, choice
  lists, providers, previews, multi-selection and shared contexts are out of scope;
  no `$` lines are emitted.
- S6 Never execute: tests compare bytes and use a pure Go model of navi's line
  grammar and entry deduplication; a lint test forbids `os/exec`, `syscall`,
  `plugin` and network imports in the package.
- S7 Preservation: consume schema-1 IR without reparsing Markdown. Frontmatter the
  `.cheat` file does not carry is preserved in a schema-1 `<sheet>.navi.json`
  companion containing complete semantic metadata, relative source path, shell
  profile and the source-ordered exported card IDs/titles/contexts. No exported
  cards means neither file. Serialization errors fail the batch before writing;
  `NAV107` is retired. See the format page for the serialization contract.
- S9 Automation: `--strict` makes any warning fail the build (`NAV000`) before
  writing. zsh, fish and nu are experimental until verified against real shells.
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
| Same parameter name declared differently in two cards | Allowed; parameters are card-local |
| Unsupported profile; `--strict` with any warning | `NAV001`; `NAV000`, nothing written |
| `<1>`, `\<name>`, `<ünï>`, `<_x>` | `NAV103` warning |
| Title empty after normalization (U+200B only) | `NAV102` warning |
| `echo "$(printf '"')"`, backtick and nested substitutions | Exported; a real open quote is still `NAV102` |
| Frontmatter including unknown nested metadata and parameter hints | Preserved in the deterministic companion; no omission warning |
| `itemHash("")`, `itemHash("abc")` | `0x1c9d3adb639f298e`, `0x64372d1c87d01e7c` |
| Header value with newline or `%` | Collapsed to one line, no new entry |
| Input order and host path | Same bytes; no absolute path in output |
| Sheet or selection with nothing exportable | `NAV106` warning; build fails if no file results |

## Diagnostics

`NAV000` (`--strict`), `NAV001`, `NAV006`, `NAV008` errors; `NAV007` retired;
`NAV101`–`NAV106` warnings; `NAV107` retired; shared
`REN001`, `IO001` and `IO003` apply. Meanings are in the format page.

## Remaining work

T009 independent exact-revision review (current increment); stronger shell syntax validation
(Bash/Zsh quote diagnostics use pinned `mvdan.cc/sh/v3/syntax` v3.14.1; Fish remains
heuristic); choice lists and shared contexts, which need their own design; nushell raw-string
analysis. A non-executing parse check against real navi, publication of generated
cheats and distribution builds are separately gated.

## Delivery evidence

Exporter/fixes reached main through [#27](https://github.com/wstein/sheetmill/pull/27)
and [#28](https://github.com/wstein/sheetmill/pull/28). Full combined review evidence
at candidate `caf1fc6` is linked from [#25](https://github.com/wstein/sheetmill/pull/25#issuecomment-6160173135).
The current sidecar increment needs its own exact-revision review and checks.
