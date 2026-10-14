---
title: grep
description: Search text files and streams for lines that match a pattern.
category: search
tags: [grep, search, regex, text]
source:
  license: EUPL-1.2
  attribution: SheetMill contributors
  note: Independently authored from the grep 2.6.0-FreeBSD (macOS) documentation; examples are displayed, never executed.
---
`grep` prints every input line that matches a pattern. It reads named files or standard input, so it works equally well at the end of a pipeline. The cards were checked against macOS grep, which aims for GNU compatibility; differences are noted.

## Basic searches {.cols-2 #basics}
### Find a word {#first-search}
```sh
grep timeout config.yaml
grep -n 'listen' nginx.conf
ps aux | grep -i postgres
```

`-n` prefixes line numbers; `-i` ignores case.

### Match precisely {#precision}
| Option | Meaning |
| --- | --- |
| `-w` | Match whole words only |
| `-x` | Match whole lines only |
| `-F` | Treat the pattern as a fixed string |
| `-v` | Print lines that do **not** match |

### Several patterns {#multiple}
```sh
grep -e error -e warning app.log
grep -f patterns.txt app.log
```

Each `-e` adds a pattern; `-f` reads one pattern per line.

### Patterns starting with a dash {.tip #dash}
```sh
grep -e '--verbose' notes.md
grep -- '-v' notes.md
```

## Regular expressions {#regex}
### Pick a syntax {#syntax}
| Option | Meaning |
| --- | --- |
| `-G` | Basic regular expressions (default) |
| `-E` | Extended: `+`, `?`, `\|` and `()` without backslashes |
| `-F` | No regex; literal text |

### Extended examples {.col-span-2 #extended}
```sh
grep -E '^(GET|POST) /api/' access.log
grep -E '[0-9]{3}-[0-9]{4}' contacts.txt
grep -E 'colou?r' style-guide.md
```

Quote patterns in single quotes so the shell leaves `|`, `*` and `$` alone. More syntax is on the [regex](regex.md) sheet.

### Print only the match {#only-matching}
```sh
grep -oE '[a-z0-9.]+@example\.org' users.csv
```

`-o` prints each matching part on its own line.

### Perl syntax {.note #perl}
macOS grep 2.6.0 has no `-P` option. Use `-E`, or a grep build that documents `-P`, for lookarounds and `\d`.

## Files and directories {#files}
### Search a tree {.col-span-2 #recursive}
```sh
grep -rn 'TODO' src/
grep -rn --include '*.go' 'context.TODO' .
grep -rn --exclude-dir vendor --exclude-dir .git 'panic(' .
```

`-r` recurses into directories; `--include` and `--exclude` take file-name globs.

### Only file names {#files-only}
| Option | Meaning |
| --- | --- |
| `-l` | Names of files with at least one match |
| `-L` | Names of files without a match |
| `-h` | Omit file names from matching lines |
| `-H` | Always show file names |

### Binary files {#binary}
```sh
grep -I -r 'secret' .
grep -a 'marker' dump.bin
```

`-I` skips binary files; `-a` treats them as text.

## Context and counting {#context}
### Surrounding lines {#around}
| Option | Meaning |
| --- | --- |
| `-A 3` | Three lines after each match |
| `-B 2` | Two lines before each match |
| `-C 2` | Two lines on both sides |

### Count and stop {#count}
```sh
grep -c 'ERROR' app.log
grep -m 1 '^version' package.toml
```

`-c` counts matching lines per file; `-m` stops after that many matches.

### Use in scripts {#quiet}
```sh
if grep -q '^ENABLE_CACHE=1$' .env; then
  echo "cache on"
fi
```

`-q` prints nothing; exit status is 0 on a match, 1 on no match and above 1 on errors.

### Highlight matches {#color}
```sh
grep --color=auto -n 'deadline' plan.md
```
