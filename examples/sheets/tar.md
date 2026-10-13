---
title: tar
description: Create, list and extract tape archives, with or without compression.
category: files
tags: [tar, archive, compression, backup]
source:
  license: EUPL-1.2
  attribution: SheetMill contributors
  note: Independently authored from the bsdtar 3.5.3 (libarchive 3.7.4) documentation; examples are displayed, never executed.
---
`tar` bundles files and directories into a single archive and unpacks them again. One mode letter picks the job (`c`, `t`, `x`, `r`, `u`) and `-f` names the archive. The cards were checked against bsdtar; GNU tar accepts the common mode letters, but confirm less common options with `tar --help`.

## Modes {#modes}
### Pick one mode {.col-span-2 #mode-letters}
| Option | Meaning |
| --- | --- |
| `-c` | Create a new archive |
| `-t` | List archive contents |
| `-x` | Extract to disk |
| `-r` | Append entries to an existing archive |
| `-u` | Append only entries newer than the archived copy |

### Name the archive {#file-option}
```sh
tar -cf site.tar public/
tar -tf site.tar
tar -cf - public/ | wc -c     # archive to stdout
```

`-f -` means standard input or output.

## Create {.cols-2 #create}
### Compressed archives {#compressed}
| Option | Meaning |
| --- | --- |
| `-z` | gzip (`.tar.gz`, `.tgz`) |
| `-j` | bzip2 (`.tar.bz2`) |
| `-J` | xz (`.tar.xz`) |
| `--zstd` | zstd (`.tar.zst`) |

```sh
tar -czf notes-2026-10.tgz notes/
```

### Let the suffix decide {#auto-compress}
```sh
tar -a -cf release.tar.xz build/
```

`-a` picks the compression from the archive name.

### Archive from another directory {#change-dir}
```sh
tar -czf backup.tgz -C /srv/app config data
```

`-C` changes directory first, so stored paths start at `config/` and `data/`.

### Leave things out {#exclude}
```sh
tar -czf src.tgz --exclude '*.log' --exclude node_modules project/
tar -czf src.tgz --exclude-vcs project/
tar -czf src.tgz -X skip-list.txt project/
```

`--exclude-vcs` skips version-control metadata such as `.git` directories.

### Archive a file list {#files-from}
```sh
tar -cf picked.tar -T list.txt
```

Combine `-T` with `--null` when the list is NUL-separated, for example from [find -print0](find.md#safe-names).

## Inspect and extract {#extract}
### List with details {#list}
```sh
tar -tvf site.tar
tar -tzf notes-2026-10.tgz 'notes/2026/*'
```

Reading auto-detects gzip, bzip2 and xz, so `-z` is optional when listing or extracting.

### Extract {#extract-basic}
| Command | Purpose |
| --- | --- |
| `tar -xf site.tar` | Unpack into the current directory |
| `tar -xf site.tar -C /tmp/site` | Unpack into an existing target directory |
| `tar -xf site.tar public/index.html` | Unpack only named entries |
| `tar -xOf site.tar public/robots.txt` | Print one entry to stdout |

### Drop a leading folder {#strip}
```sh
tar -xzf tool-1.4.0.tgz --strip-components 1 -C ~/opt/tool
```

Removes the first path element; entries with fewer elements are skipped.

### Keep existing files {#keep-old}
```sh
tar -xkf site.tar
```

`-k` refuses to overwrite files already on disk.

### Untrusted archives {.warning #untrusted}
Extraction overwrites files with the same name. List an unknown archive with `-tvf` first and extract into an empty directory; never add `-P`, which keeps absolute paths and `..` components.

## Variants {#variants}
### Rename while archiving {.note #rename}
```sh
tar -cf out.tar -s '/^build/app-1.4/' build/
```

`-s` path substitution is a bsdtar option; check your GNU tar manual for its equivalent.

### Check your implementation {#version}
```sh
tar --version
```

macOS ships bsdtar (libarchive); most Linux distributions ship GNU tar. Option spellings can differ beyond the basics.
