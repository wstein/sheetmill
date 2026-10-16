---
title: find
description: Walk directory trees and select files by name, type, size, age or permissions, then act on them.
category: search
tags: [find, files, filesystem, shell]
cheatsheet:
  updated: "2026-10-14"
source:
  license: EUPL-1.2
  attribution: SheetMill contributors
  note: Independently authored from the macOS (BSD) find manual page; examples are displayed, never executed.
---
`find` walks one or more starting paths and evaluates an expression for every entry. Tests filter, actions do something with the survivors; the default action is `-print`. Always give a starting path: BSD `find` requires one.

## Match by name and type {.cols-2 #names}
### Names and paths {#by-name}
| Test | Meaning |
| --- | --- |
| `-name '*.md'` | Last path component matches a glob |
| `-iname 'readme*'` | Same, ignoring case |
| `-path '*/docs/*'` | Whole path matches a glob |

```sh
find . -name '*.go' -not -name '*_test.go'
```

Quote globs so the shell does not expand them first.

### Entry types {#by-type}
| Test | Meaning |
| --- | --- |
| `-type f` | Regular file |
| `-type d` | Directory |
| `-type l` | Symbolic link |
| `-empty` | Empty file or directory |

### Limit depth {#depth}
```sh
find . -maxdepth 1 -type d
find src -mindepth 2 -name '*.json'
```

### Combine tests {#operators}
```sh
find . \( -name '*.jpg' -o -name '*.png' \) -size +1M
find . -type f ! -perm -u=w
```

Tests are joined by an implicit AND; `-o` is OR and `!` or `-not` negates. Escape parentheses for the shell.

## Size, time and ownership {#attributes}
### Size {#by-size}
| Test | Meaning |
| --- | --- |
| `-size +100M` | Larger than 100 MiB |
| `-size -10k` | Smaller than 10 KiB |
| `-size 0` | Zero length |

### Modification time {.col-span-2 #by-time}
| Test | Meaning |
| --- | --- |
| `-mtime -7` | Modified less than 7 days ago |
| `-mtime +30` | Modified more than 30 days ago |
| `-mmin -15` | Modified within the last 15 minutes |
| `-newer stamp.txt` | Modified after `stamp.txt` |

`+n` means more than, `-n` less than, plain `n` exactly n units.

### Owner and permissions {#by-owner}
```sh
find /srv -user deploy -type f
find . -type f -perm -o=w
```

`-perm -mode` matches when all listed bits are set.

## Act on results {#actions}
### Run a command {.col-span-2 #exec}
```sh
find . -name '*.md' -exec wc -l {} +
find . -name '*.orig' -exec rm {} \;
find logs -type f -execdir gzip {} +
```

`{} +` passes many paths per call; `{} \;` runs once per file. `-execdir` runs from each file's directory.

### Names with spaces {.tip #safe-names}
```sh
find . -type f -name '*.txt' -print0 | xargs -0 grep -l TODO
```

NUL separators survive spaces and newlines in file names. See [grep](grep.md#files-only) for the `-l` option.

### Delete matches {.warning #delete}
```sh
find build -name '*.tmp' -print     # preview first
find build -name '*.tmp' -delete
```

`-delete` is irreversible and acts the moment it is evaluated, so keep it last and preview with `-print` first.

## Shape the walk {#walk}
### Skip a directory {#prune}
```sh
find . -name .git -prune -o -type f -print
```

`-prune` stops descent into matching directories; the explicit `-print` keeps them out of the output.

### Follow symbolic links {.note #symlinks}
```sh
find -L assets -type f
```

By default links are not followed. `-L` follows them and tests the target instead.

### Regular expressions {#regex}
```sh
find -E . -regex '.*/v[0-9]+\.[0-9]+/.*'
```

`-regex` matches the whole path. `-E` selects extended syntax on BSD find; GNU find uses `-regextype` instead. See [regex](regex.md) for syntax.
