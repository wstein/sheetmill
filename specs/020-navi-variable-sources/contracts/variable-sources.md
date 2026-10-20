# Public contract - 020

The user-facing pages (format docs, schema, diagnostics) carry the same contract
once implemented (tasks T010, T011, T014). Until then this file is normative.

## Source syntax

````markdown
---
title: Git
license: EUPL-1.2
tags: [vcs]
cheatsheet:
  tool: git
  cards:
    fetch-branch:
      params:
        remote:
          type: string
          description: Remote to fetch from
          source: git remote
          options: {column: 1}
        branch:
          type: string
          description: Branch to fetch
          choices: [main, "it's", süd]
    delete-branches:
      params:
        branches:
          type: string
          source: git branch --format='%(refname:short)'
          options: {multi: true}
---
## Branches

### Fetch a branch {#fetch-branch .command}

```bash
git fetch <remote> <branch>
```

### Delete branches {#delete-branches .command}

```bash
git branch -d <branches>
```
````

`cheatsheet.navi.cards.<id>.params.<name>` accepts the same fields with the same
rules (errors are `NAV006`). A parameter without `choices` and `source` is free input
as before.

## Metadata schema addition

`schema/cheatsheet.schema.json`, under `cheatsheet.cards.*.params.*.properties`:

```json
"choices": {
  "type": "array", "minItems": 1, "maxItems": 64, "uniqueItems": true,
  "items": {"type": "string", "minLength": 1, "maxLength": 128}
},
"source": {"type": "string", "minLength": 1},
"options": {
  "type": "object",
  "properties": {
    "column": {"type": "integer", "minimum": 1, "maximum": 255},
    "delimiter": {"type": "string", "minLength": 1},
    "header_lines": {"type": "integer", "minimum": 0, "maximum": 255},
    "multi": {"type": "boolean"},
    "prevent_extra": {"type": "boolean"}
  },
  "additionalProperties": false
}
```

plus `"not": {"required": ["choices", "source"]}` on the parameter. Character rules,
`type: int` choices, option/source dependencies and `multi` with `prevent_extra`
are checked by source validation (the schema comments say so, as for `description`).

## Output line

```
$ <name>: <command>[ --- <option> [<option> ...]]
```

- One line; no newline, control, format character, U+2028/U+2029 or U+2800 inside.
- Placed directly after the card's command lines; no blank line between them.
- `<name>` matches `[a-z][a-z0-9_]{0,31}`, is declared, appears in the command, and
  has one line per card. Order: first placeholder use in the command.
- Options in order `--column <n>`, `--delimiter '<s>'`, `--header-lines <n>`,
  `--multi`, `--prevent-extra`; `<s>` is written in single quotes.

## Profile quoting for `choices`

| Profile | Command | Literal quoting |
| --- | --- | --- |
| `bash`, `zsh` | `printf '%s\n' <lit> ...` | wrap in `'`; each `'` becomes `'\''` |
| `fish` | `printf '%s\n' <lit> ...` | wrap in `'`; `\` becomes `\\`, `'` becomes `\'` |
| `nu` | `[ <lit> ... ] \| to text` | wrap in `"`; `\` becomes `\\`, `"` becomes `\"` |

Literals are separated by one space and are never reordered, merged or trimmed.
Newlines and controls cannot occur (rejected earlier). nu output and leading-dash
literals under fish/zsh are listed as to-be-measured in [research](../research.md);
the nu and fish profiles stay experimental as in 006.

## Golden examples

Source: the sheet above at `sheets/git.md`. Header lines as in the
[navi format page](../../../docs/format/navi-profile.md) (only the `Shell` value
and the fence language differ between profiles). Entries for `bash`:

```text
; card: fetch-branch
% sheetmill:sheets/git.md#fetch-branch!, git, vcs
# Fetch a branch
git fetch <remote> <branch>
$ remote: git remote --- --column 1
$ branch: printf '%s\n' 'main' 'it'\''s' 'süd'

; card: delete-branches
% sheetmill:sheets/git.md#delete-branches!, git, vcs
# Delete branches
git branch -d <branches>
$ branches: git branch --format='%(refname:short)' --- --multi
```

`zsh` (fence `zsh`): identical entries. `fish` (fence `fish`; `source` written for
fish): the `branch` line is

```text
$ branch: printf '%s\n' 'main' 'it\'s' 'süd'
```

`nu` (fence `nu`; `source` written for nu): the `branch` line is

```text
$ branch: [ "main" "it's" "süd" ] | to text
```

## Diagnostics

| Code | Severity | Meaning |
| --- | --- | --- |
| `CMD001` | error | (extended) invalid `choices`, `source` or `options`; `source` with `choices`; option without source/choices |
| `NAV006` | error | (extended) the same rules in the legacy namespace; conflicting declarations including the new fields |
| `NAV108` | warning | Card skipped: a `source` or option string cannot be written as one valid `$` line (`---` in a source, a navi `<name>` token in a source, a trailing backslash, an unterminated quote, or `---` or `'` in an option string) |
| `NAV109` | warning | Card skipped: a choice cannot be carried by navi's line grammar (`---` or a navi `<name>` token) |
| `NAV104` | warning | (unchanged) a declared parameter is unused or a placeholder is undeclared |

Source diagnostics point at the nested key; direct IR uses the `1:1` anchor (015).
`NAV108`/`NAV109` use the code block span of the card. `--strict` turns them into
`NAV000` before writing. Nothing is rewritten to make a card pass.

## Negative table (expected results)

| Input | Result |
| --- | --- |
| `choices: []` | `CMD001`, nothing written |
| `choices: [a, a]` | `CMD001` |
| `choices: [1, 2]` (numbers) | `CMD001`, no coercion |
| `choices: ["a\nb"]`, `["\t"]`, `[""]`, `[" a"]` | `CMD001` |
| `type: int`, `choices: ["x"]` | `CMD001` |
| `source` and `choices` both set | `CMD001` |
| `source: "a\nb"`, `source: ""` | `CMD001` |
| `options: {column: 0}`, `{column: "2"}`, `{foo: 1}` | `CMD001` |
| `options: {multi: true}` alone | `CMD001` |
| `options: {column: 1}` with `choices` | `CMD001` |
| `options: {multi: true, prevent_extra: true}` | `CMD001` |
| `source: echo a --- b` | `NAV108`, card skipped |
| `source: echo <other>` | `NAV108` (dependent variables out of scope) |
| `source: echo a \` | `NAV108` |
| `source: echo "a` (bash/zsh/fish) | `NAV108` |
| `options: {delimiter: "it's"}` | `NAV108` |
| `options: {delimiter: "a---b"}` | `NAV108` (navi splits the line on `---`) |
| `options: {map: "x"}` | `CMD001` (`map` is out of scope) |
| `choices: ["a---b"]`, `["<x>"]` | `NAV109` |
| parameter declared, never used | `NAV104` (existing) |
| `choices: [a'b, "c\\d", "$HOME", "%s"]` | exported, quoted per profile |
