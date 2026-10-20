# Source evidence - 020

Reviewed 2026-10-20 by reading documentation and source only. No navi, shell,
finder or example command was run. The version and commit are those of
[006 research](../006-navi-renderer/research.md); links below use that immutable
commit (`5515367dc8a2a561d2d82b8352f729b1da4120d3`). The syntax page at that commit
(`docs/cheatsheet_syntax.md`) was compared with the current master page
(`docs/cheatsheet/syntax/README.md`); the `$` line, `---` options and the supported
option list are the same in both.

## Verified from source

| Upstream source | Observed behavior and effect on the design |
| --- | --- |
| [Syntax](https://github.com/denisidoro/navi/blob/5515367dc8a2a561d2d82b8352f729b1da4120d3/docs/cheatsheet_syntax.md) | `$ name: command` lines define suggestions for `<name>`; `---` separates fzf/navi options. Documented options: `--column`, `--map` (experimental; not exported, see below), `--prevent-extra` (experimental), `--fzf-overrides` (experimental), `--expand` (experimental), and forwarded `--multi`, `--header-lines`, `--delimiter`, `--query`, `--filter`, `--header`, `--preview`, `--preview-window`. |
| [Parser](https://github.com/denisidoro/navi/blob/5515367dc8a2a561d2d82b8352f729b1da4120d3/src/parser.rs) `VAR_LINE_REGEX` | `^\$\s*([^:]+):(.*)`: name is everything up to the first colon (trimmed), command is the rest. SheetMill names contain no colon. |
| Parser, `parse_variable_line` | The command text is split on every `---`; the first part is the command, the second the options, further parts are ignored. `---` therefore cannot appear inside a command or a quoted literal: the exporter rejects it (`NAV108`/`NAV109`). |
| Parser, `read_lines` | A line is a variable line when it starts with `$` and contains `:` outside a Markdown fence. A trailing `\` continues the line, so a trailing backslash is rejected. Reading a variable line first writes the pending entry and clears its snippet; the entry's tags are the last `%` line, so `$` lines belong directly after the card's command. A blank line before a `$` line would append a separator to the snippet, so none is emitted. |
| Parser, `parse_opts` | Options are split with `shellwords::split` (independent of the configured shell). `--multi`, `--prevent-extra`, `--expand` are flags; other options are flag/value pairs. `--column` and `--header-lines` (alias `--headers`) parse as `u8`; unknown flags are ignored but still consume a value. `multi` wins over `prevent-extra`. An unclosed quote is an error. |
| [Variable map](https://github.com/denisidoro/navi/blob/5515367dc8a2a561d2d82b8352f729b1da4120d3/src/structures/cheat.rs) | Suggestions are keyed by the hash of the tags line and the variable name. SheetMill's unique context tag makes them card-local. |
| [Actor](https://github.com/denisidoro/navi/blob/5515367dc8a2a561d2d82b8352f729b1da4120d3/src/commands/core/actor.rs) | The suggestion command is passed to the configured shell with `-c`; its stdout is the suggestion list. Before running it, navi substitutes placeholders matching `\\?<(\w[\w\d\-_]*)>` inside the command (implicit dependencies). A `<name>` inside a literal or source would therefore create a dependency: dependent variables are out of scope, so such text is rejected. Only the snippet, not the suggestion command, gets the `navi ` executable rewrite. |
| [Item identity](https://github.com/denisidoro/navi/blob/5515367dc8a2a561d2d82b8352f729b1da4120d3/src/structures/item.rs) | The pinned hash input is tags plus description only. `$` lines do not change entry hashes. |

## Decisions

- Export-only. `.cheat` ingestion is not planned; the sheet stays the single source.
- Target-neutral shape (`choices`, `source`) is validated by core (`CMD001`); navi
  grammar limits (`---`, tokens, backslash, quotes) are navi export warnings
  (`NAV108`, `NAV109`), consistent with 006's skip-with-warning rule.
- Options are a closed list limited to what the parser section above shows and that
  needs no quoting beyond single quotes: `column`, `delimiter`, `header_lines`,
  `multi`, `prevent_extra`. `map` is out of scope: it runs navi's `_navi_input` and
  `_navi_map_fn` helper functions in non-fish shells, is experimental, and cannot work
  under nu. Option strings are written in single quotes. A value containing `---` is
  rejected (navi splits the whole remainder on it); a value containing `'` is rejected
  for simplicity, although the splitter (`'([^']*)'` plus `\\.?`, adjacent pieces joined)
  would parse `'it'\''s'`.
- Choices become `printf '%s\n' ...` in bash and zsh, the same in fish, and a quoted
  list piped to `to text` in nu. `$` line order follows first use in the command.
- Header, tags, contexts and companion schema are unchanged.

## Answered by docs and source

1. `shellwords::split` keeps backslashes inside single quotes verbatim (navi's own
   docs use `--delimiter '\s\s+'`).
2. fzf `--multi` output is newline-joined, the last character is truncated, and the
   text is inserted raw unless `--expand` is given (`src/finder/post.rs`).
3. navi never trims suggestion lines; fzf's own handling is not documented. Choices
   with surrounding whitespace stay rejected.
4. navi runs `<configured shell words> -c <command>` (`src/common/shell.rs`).

## To be measured (not verified; no claim is made)

1. A choice starting with `-` after the `printf` format in fish and zsh (bash builtin
   semantics are documented; fish/zsh are not verified).
2. The exact output of nu `to text` on a list of strings (trailing newline).

Both are optional, separately authorized measurements and block nothing. Re-check
the upstream source when moving to another navi version.
