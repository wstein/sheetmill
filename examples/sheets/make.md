---
title: GNU Make
description: Describe how files are rebuilt from their prerequisites and let GNU Make run only the recipes that are out of date.
category: build
tags: [make, gnu, build, makefile, automation]
source:
  license: EUPL-1.2
  attribution: SheetMill contributors
  note: Independently authored from the GNU Make 3.81 documentation; examples are displayed, never executed.
---
make compares file modification times and reruns a rule's recipe when its target is missing or older than a prerequisite. This sheet sticks to features present in GNU Make 3.81; newer releases add more.

## Rules {.cols-2 #rules}
### Anatomy of a rule {#anatomy}
```make
report.pdf: report.md style.css
	pandoc report.md --css style.css -o report.pdf
```

The target comes before the colon and prerequisites after it. Each recipe line runs in its own shell.

### Recipe lines need a tab {.warning #tabs}
```make
clean:
	rm -f *.o app
```

Recipe lines must begin with a real tab character. Spaces cause a "missing separator" error in 3.81.

### Phony targets {#phony}
```make
.PHONY: all test clean

all: app
test: app
	./run-tests.sh
clean:
	$(RM) app *.o
```

Mark targets that are not files as phony, so a stray file named `test` never makes them look up to date.

### Recipe prefixes {#prefixes}
| Prefix | Effect |
| :-- | :-- |
| `@echo done` | Run without echoing the command |
| `-rm stale.lock` | Ignore a failing exit status |
| `$$HOME` | Pass a literal `$` to the shell |

### Default and order-only {#goals}
```make
.DEFAULT_GOAL := all

build/app.o: app.c | build
	$(CC) -c app.c -o $@
build:
	mkdir -p build
```

Without `.DEFAULT_GOAL`, make builds the first target in the file. Prerequisites after `|` must exist but never trigger a rebuild by being newer.

## Variables {.cols-2 #variables}
### Assignment flavors {#assignment}
| Form | Meaning |
| :-- | :-- |
| `SRC = $(wildcard *.c)` | Recursive: expanded each time it is used |
| `NOW := $(shell date +%s)` | Simple: expanded once, at definition |
| `CC ?= cc` | Set only if not already defined |
| `CFLAGS += -Wall` | Append with a space |
| `make CFLAGS=-O2` | Command-line value overrides the makefile |

### Automatic variables {.col-span-2 #automatic}
| Variable | Value inside a recipe |
| :-- | :-- |
| `$@` | The target being built |
| `$<` | The first prerequisite |
| `$^` | All prerequisites, duplicates removed |
| `$+` | All prerequisites, duplicates kept |
| `$?` | Prerequisites newer than the target |
| `$*` | The stem matched by `%` |
| `$(@D)`, `$(@F)` | Directory and file part of `$@` |

### Built-in variables {#builtin-vars}
| Variable | Typical built-in default |
| :-- | :-- |
| `CC` | `cc` |
| `RM` | `rm -f` |
| `SHELL` | `/bin/sh` |
| `MAKE` | The running make, for recursion |
| `CURDIR` | The working directory |

Print the full built-in database with `make -p -f /dev/null`.

### Per-target values and export {#target-vars}
```make
debug: CFLAGS += -g -O0
debug: app

export LANG := C
```

A target-specific value applies while that target and its prerequisites are built. `export` passes a variable to recipe shells.

## Patterns and functions {.cols-2 #patterns}
### Pattern rules {.col-span-2 #pattern-rules}
```make
OBJS := main.o parse.o util.o

app: $(OBJS)
	$(CC) $(LDFLAGS) $^ -o $@

%.o: %.c defs.h
	$(CC) $(CFLAGS) -c $< -o $@
```

`%` matches the same stem on both sides. A built-in `%.o: %.c` rule already exists; `make -r` disables built-in rules.

### Text functions {#functions}
| Call | Result |
| :-- | :-- |
| `$(wildcard src/*.c)` | Existing files matching a glob |
| `$(patsubst %.c,%.o,$(SRC))` | Pattern replacement per word |
| `$(SRC:.c=.o)` | Suffix substitution shorthand |
| `$(addprefix build/,$(OBJS))` | Prefix every word |
| `$(filter %.c,$(FILES))` | Keep matching words |
| `$(filter-out test_%,$(SRC))` | Drop matching words |
| `$(notdir src/a.c)`, `$(dir src/a.c)` | `a.c` and `src/` |

### Loops and shell output {#foreach}
```make
DIRS := lib cli docs
STAMPS := $(foreach d,$(DIRS),$(d)/.stamp)
REV := $(shell git rev-parse --short HEAD)
```

`$(shell ...)` runs while the makefile is read; pair it with `:=` so it runs once. Commit IDs are covered on the [git sheet](git.md).

## Structure and conditionals {.cols-2 #structure}
### Conditionals {#conditionals}
```make
ifeq ($(OS),Windows_NT)
  EXE := .exe
else
  EXE :=
endif

ifndef PREFIX
  PREFIX := /usr/local
endif
```

`ifneq` and `ifdef` complete the set. Conditionals are evaluated while the makefile is read, not when recipes run.

### Include and fail early {#include}
```make
include config.mk
-include $(OBJS:.o=.d)

ifeq ($(strip $(VERSION)),)
  $(error VERSION is not set)
endif
```

`-include` ignores missing files. `$(error ...)` stops make with a message.

### Recursive make {#recursive}
```make
SUBDIRS := lib cli

all:
	for d in $(SUBDIRS); do $(MAKE) -C $$d || exit 1; done
```

Use `$(MAKE)` rather than `make` so options such as `-n` and `-j` reach the child.

## Running make {.cols-2 #running}
### Common options {.col-span-2 #options}
| Option | Effect |
| :-- | :-- |
| `-n` | Print the recipes without running them |
| `-j 4` | Run up to four jobs at once |
| `-k` | Keep building other targets after an error |
| `-B` | Rebuild every target unconditionally |
| `-C dir` | Change to `dir` before reading makefiles |
| `-f build.mk` | Use another makefile |
| `-s` | Do not echo recipes |
| `-q` | Run nothing; exit status tells if up to date |

### Debugging a build {#debugging}
```sh
make -n install
make --debug=b app
make -W parse.c -n
make --warn-undefined-variables
```

`-W file` pretends a file just changed; with `-n` it previews the rebuild.

### Exit status {.note #exit-status}
| Status | Meaning |
| :-- | :-- |
| `0` | All makefiles parsed and targets built |
| `1` | With `-q`, something needs rebuilding |
| `2` | An error occurred |

Without `-f`, make looks for `GNUmakefile`, `makefile`, then `Makefile`.
