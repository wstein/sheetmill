---
title: jq
description: Filter, reshape and summarize JSON on the command line with jq's small expression language.
category: data
tags: [jq, json, data, cli, filters]
source:
  license: EUPL-1.2
  attribution: SheetMill contributors
  note: Independently authored from the jq 1.8.2 documentation; examples are displayed, never executed.
---
A jq program is a filter: it takes each JSON input and emits zero or more JSON outputs. The examples assume an input such as `{"team":"ops","members":[{"name":"Ada","age":36,"tags":["oncall"]}]}`.

## Running jq {.cols-2 #running}
### Command-line options {.col-span-2 #options}
| Option | Effect |
| :-- | :-- |
| `-r` | Print strings without JSON quotes |
| `-c` | One compact JSON value per line |
| `-n` | Start from `null` instead of reading input |
| `-s` | Read every input into one array |
| `-R` | Treat each input line as a string |
| `-e` | Exit `1` if the last output is `false` or `null` |
| `-S` | Sort object keys in the output |
| `-j` | Raw output without trailing newlines |
| `--tab`, `--indent 4` | Choose the pretty-print indentation |

### Pass values in safely {#arguments}
```sh
jq --arg who "$USER" '.members[] | select(.name == $who)' team.json
jq --argjson limit 30 '.members[] | select(.age > $limit)' team.json
jq -n --rawfile body notes.txt '{text: $body}'
jq -n '$ENV.HOME'
```

`--arg` always binds a string; use `--argjson` for numbers, booleans or objects.

### Exit status in scripts {.tip #exit-status}
```sh
if jq -e '.members | length > 0' team.json >/dev/null; then
  echo "team has members"
fi
```

With `-e`, a run that produces no valid result at all exits with `4`.

## Reading values {.cols-2 #reading}
### Paths and indexes {#paths}
| Filter | Result |
| :-- | :-- |
| `.` | The whole input, pretty-printed |
| `.team` | Value of one key |
| `.members[0].name` | Nested lookup |
| `.members[-1]` | Last array element |
| `.members[1:3]` | Array slice, end exclusive |
| `.["odd key"]` | Key that is not a plain identifier |
| `.team?` | Same lookup, but suppress type errors |

### Iterate and combine {#iterate}
```jq
.members[]              # one output per element
.members[].name         # names, one per line
.team, .members[0].age  # comma emits both results
.members[] | .tags[]    # pipe feeds each result onward
..                      # every value, recursively
```

### Defaults and missing data {#defaults}
```jq
.owner // "unassigned"
.members[].email // empty
try (.raw | fromjson) catch "bad payload"
```

`a // b` yields `b` when `a` produces only `null`, `false` or nothing.

## Building output {.cols-2 #building}
### Construct new values {.col-span-2 #construct}
```jq
{team, count: (.members | length)}
{(.team): [.members[].name]}
[.members[] | {name, senior: (.age >= 40)}]
"\(.team) has \(.members | length) people"
```

`{team}` is shorthand for `{team: .team}`; a parenthesized key is computed. Ready-made JSON can go straight to [curl --json](curl.md#post-json).

### Format strings {#formats}
| Filter | Output |
| :-- | :-- |
| `@csv` | Array as a CSV row, strings quoted |
| `@tsv` | Array as a tab-separated row |
| `@sh` | Shell-quoted string or words |
| `@uri` | Percent-encoded text |
| `@base64`, `@base64d` | Encode or decode Base64 |
| `@json`, `@html` | JSON text or HTML-escaped text |

### Rows for spreadsheets {#rows}
```sh
jq -r '.members[] | [.name, .age] | @csv' team.json
jq -r '"name\tage", (.members[] | [.name, .age] | @tsv)' team.json
```

Combine with `-r`, otherwise each row is printed as a quoted JSON string.

## Transforming data {.cols-2 #transforming}
### Filter and map {#select-map}
```jq
.members | map(select(.age < 30))
.members | map(.name | ascii_upcase)
[.members[] | select(any(.tags[]; . == "oncall"))]
.members | any(.age > 60)
```

### Sort, group, count {#aggregate}
| Filter | Purpose |
| :-- | :-- |
| `sort_by(.age)` | Order objects by a field |
| `group_by(.team)` | Arrays of objects sharing a value |
| `unique_by(.name)` | One object per distinct value |
| `min_by(.age)`, `max_by(.age)` | Extreme element |
| `length` | Size of array, object or string |
| `add` | Sum numbers, join strings, concatenate arrays |
| `keys` | Sorted object keys |

### Update in place {#update}
```jq
.members[].age += 1
.members[].name |= ascii_downcase
.team = "platform"
del(.members[] | select(.age > 65))
with_entries(.key |= "x_" + .)
```

`|=` runs the right side on the old value; plain `=` evaluates the right side against the whole input.

### Accumulate with reduce {#reduce}
```jq
reduce .members[] as $m (0; . + $m.age)
reduce .members[] as $m ({}; .[$m.name] = $m.age)
```

The first argument is the starting value; the second updates it once per item.

## Strings and regular expressions {.cols-2 #strings}
### String helpers {#string-helpers}
| Filter | Purpose |
| :-- | :-- |
| `split(",")` | String to array on a literal separator |
| `join("-")` | Array of strings to one string |
| `ltrimstr("v")` | Remove a prefix if present |
| `trim` | Remove surrounding whitespace |
| `startswith("api")` | Prefix test |
| `tostring`, `tonumber` | Convert between types |

### Regex functions {#regex}
```jq
.name | test("^a"; "i")
.version | capture("(?<major>[0-9]+)\\.(?<minor>[0-9]+)")
.text | [scan("[0-9]+")]
.path | gsub("/+"; "/")
```

jq uses Oniguruma's Perl-style syntax, not POSIX ERE; backslashes are doubled inside JSON strings. See [the regex sheet](regex.md#flavors).

### Dates {.note #dates}
```jq
now | todate
"2026-03-01T08:00:00Z" | fromdate
.ts | gmtime | strftime("%Y-%m-%d")
```

Date builtins work in UTC: `fromdate` reads ISO 8601 text, `todate` and `gmtime` take Unix seconds.
