---
title: Regular expressions
description: POSIX extended regular expressions as used by grep -E, with notes on where Perl-style engines differ.
category: text
tags: [regex, posix, ere, grep, patterns]
source:
  license: EUPL-1.2
  attribution: SheetMill contributors
  note: Independently authored from the macOS re_format(7) POSIX 1003.2 documentation and grep option documentation; examples are displayed, never executed.
---
This sheet describes **POSIX extended regular expressions (ERE)**, the flavor selected by `grep -E`. Unless a card says otherwise, the syntax is portable ERE; Perl-style differences are collected in the last section.

## Building blocks {.cols-2 #basics}
### Ordinary and special characters {#special-chars}
| Pattern | Matches |
| :-- | :-- |
| `cat` | The literal text `cat` |
| `.` | Any single character |
| `\.` | A literal dot |
| `\\` | A literal backslash |
| `a\+b` | Literal `a+b` |

Outside brackets these need a backslash to be literal: `^ . [ $ ( ) | * + ? { \`.

### Anchors {#anchors}
| Pattern | Matches |
| :-- | :-- |
| `^Error` | `Error` at the start of a line |
| `;$` | `;` at the end of a line |
| `^$` | An empty line |
| `^[0-9]+$` | A line made only of digits |

### Alternation and grouping {#groups}
```text
red|green|blue          one of three words
gr(a|e)y                gray or grey
(ab)+                   ab, abab, ababab, ...
^(GET|POST) /api/       request lines for two methods
```

`|` has the lowest precedence, so `^a|b$` means `^a` or `b$`; group to change that.

## Repetition {.cols-2 #repetition}
### Quantifiers {#quantifiers}
| Piece | Repeats the atom |
| :-- | :-- |
| `x*` | Zero or more times |
| `x+` | One or more times |
| `x?` | Zero or one time |
| `x{3}` | Exactly three times |
| `x{2,}` | Two or more times |
| `x{2,5}` | Two to five times |

A quantifier applies to the single atom before it: a character, a bracket expression or a group.

### Bounds in practice {#bounds}
```text
[0-9]{4}-[0-9]{2}-[0-9]{2}        a date such as 2026-10-14
[A-Fa-f0-9]{6}                    six hex digits
(na){2,}                          nana, nanana, ...
```

A `{` that is not followed by a digit is an ordinary character on macOS. Bounds may not exceed 255 there.

### Leftmost, then longest {.note .col-span-2 #longest}
```text
pattern:  cat|category
input:    category
POSIX:    category        (longest match at the earliest position)
Perl/JS:  cat             (first alternative that succeeds)
```

POSIX engines choose the earliest starting position, then the longest match from there. Backtracking engines stop at the first alternative that works.

## Bracket expressions {.cols-2 #brackets}
### Sets and ranges {#sets}
| Pattern | Matches one character |
| :-- | :-- |
| `[aeiou]` | Any listed vowel |
| `[^aeiou]` | Anything except those |
| `[a-z]` | A range in the collating order |
| `[]x]` | `]` or `x`; put `]` first |
| `[a-]` | `a` or `-`; put `-` first or last |
| `[*.]` | A literal star or dot |

Inside brackets, `\` and most other special characters are ordinary.

### Character classes {#classes}
| Class | Members |
| :-- | :-- |
| `[[:digit:]]` | Decimal digits |
| `[[:alpha:]]` | Letters |
| `[[:alnum:]]` | Letters and digits |
| `[[:space:]]` | Whitespace, including tab and newline |
| `[[:upper:]]`, `[[:lower:]]` | Upper- or lowercase letters |
| `[[:punct:]]` | Punctuation |
| `[[:xdigit:]]` | Hexadecimal digits |

Classes are locale-aware and only valid inside a bracket expression, so write `[[:digit:]]`, not `[:digit:]`.

### Ranges are locale-dependent {.warning #ranges}
```text
[[:lower:]]+       prefer a class
[a-z]+             may differ between locales
```

Ranges follow the collating sequence, so prefer named classes in portable patterns.

## Matching with grep {.cols-2 #grep}
### Useful options {#grep-options}
| Option | Effect |
| :-- | :-- |
| `-E` | Use extended (ERE) syntax |
| `-i` | Ignore case |
| `-v` | Select lines that do not match |
| `-o` | Print only the matched parts |
| `-w`, `-x` | Match whole words or whole lines |
| `-c`, `-n` | Count matches or show line numbers |
| `-q` | No output; exit status only |

More search patterns live on the [grep sheet](grep.md).

### Everyday patterns {#everyday}
```sh
grep -E '^[[:space:]]*#' app.conf
grep -Ev '^[[:space:]]*(#|$)' app.conf
grep -Eo '[0-9]{1,3}(\.[0-9]{1,3}){3}' access.log
grep -Ei 'warn(ing)?|error' build.log
```

Quote patterns in single quotes so the shell leaves `$`, `*` and `|` alone.

### Basic vs extended syntax {#bre-ere}
| ERE (`grep -E`) | BRE (`grep`) |
| :-- | :-- |
| `a+`, `a?` | `aa*`, `a\{0,1\}` |
| `(ab)` | `\(ab\)` |
| `a{2,3}` | `a\{2,3\}` |
| `a\|b` | no portable equivalent |

Back-references such as `\1` belong to BRE; strict ERE has none.

## Other flavors {.cols-2 #flavors}
### Perl-style shortcuts {.col-span-2 #shortcuts}
| Perl-style (PCRE, JavaScript) | POSIX ERE spelling |
| :-- | :-- |
| `\d`, `\D` | `[[:digit:]]`, `[^[:digit:]]` |
| `\s`, `\S` | `[[:space:]]`, `[^[:space:]]` |
| `\w`, `\W` | `[[:alnum:]_]`, `[^[:alnum:]_]` |
| `\b` | no ERE equivalent; try `grep -w` |

macOS accepts these shortcuts only in its "enhanced" mode, and spells word edges `[[:<:]]` and `[[:>:]]`.

### Features without an ERE equivalent {#perl-only}
```text
(?:abc)          group without capturing
a+?  a*?         lazy (shortest) repetition
(?=next)         lookahead
(?<year>\d{4})   named group
(?i)abc          inline case-insensitive flag
```

These belong to backtracking engines such as PCRE, JavaScript or the Oniguruma engine used by [jq](jq.md#regex). Test such patterns in the tool that will run them.
