# Data model - 020

## Source metadata (canonical)

Existing: `cheatsheet.cards.<id>.params.<name>` with required `type`
(`string|int|path`) and optional `description`; legacy alias
`cheatsheet.navi.cards.<id>.params.<name>`. 020 adds, to the same closed mapping:

| Field | Type | Rules |
| --- | --- | --- |
| `choices` | list of strings | 1 to 64 items; each 1 to 128 code points; no controls, format characters, U+2028/U+2029, U+2800, surrounding whitespace; no duplicates; `type: int` requires `^-?[0-9]+$` |
| `source` | string | nonempty, single line, same character rules; exclusive with `choices` |
| `options` | mapping | only the keys below; requires `source` or `choices` |

`options` keys: `column` (integer 1 to 255, `source` only), `delimiter` (string, `source`
only), `header_lines` (integer 0 to 255, `source` only), `multi`
(boolean), `prevent_extra` (boolean). Strings follow the `source` character rules.
`multi` and `prevent_extra` are not both true. Integers are canonical decoded integers;
strings, booleans and floats are never coerced (as in 015).

## IR

The schema-1 IR is **unchanged**: parameters live in the generic `Document.Metadata`
map (015 reads them nonmutatingly), so no field, class or schema number is added and
the parser goldens and IR bytes of every existing sheet stay identical. Hugo
output and `.navi.json` companions need no change: the companion's `metadata`
already carries the mapping verbatim (sorted keys), so a sheet using the fields has a
companion that differs only by that metadata.

## Derived export model (renderer-internal)

`param` gains `Choices []string`, `Source string` and `Options` (typed, with presence
flags for booleans and integers). Equality of two declarations (old/new namespaces,
015) compares all fields, so `NAV006` conflicts include them; absent and empty
`options` are the same.

## Output

For a card with N used parameters having a variable source, N `$` lines follow the
command lines in first-use order:

```
$ <name>: <command>
$ <name>: <command> --- <option> [<option> ...]
```

Option order is fixed: `--column`, `--delimiter`, `--header-lines`,
`--multi`, `--prevent-extra`. Booleans false are omitted. String values are single
quoted and must not contain `---` or `'`. `choices` produce `<command>` from the profile quoting table in the
[contract](contracts/variable-sources.md); `source` is copied.

## Identity

The pinned prehash stays `tagLine + oneLine(title)`. `$` lines are not in it, so
entry hashes, contexts and NAV008 behavior are unchanged for every sheet. Schema
versions: IR 1, companion 1, `.cheat` format 1 (unchanged).

## Line invariants (checked by tests)

One `$` entry is one line; no blank line inside a card; no line other than `%`, `#`,
`$`, `;` and command lines; every `$` name is declared, used and appears once per
card; every placeholder with a declaration that has a variable source has exactly one
`$` line; column-0 `%`, `@`, `;`, `$` appear only as the generator's own markers.
