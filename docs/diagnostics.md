---
title: "Diagnostics"
description: "Every SheetMill diagnostic code, what it means, how to fix it, and how to check sheets in CI."
category: reference
---
Every message has the form `path:line:column: CODE: message` on stderr. Errors
stop the command with exit status 1 (invalid source) or 2 (usage or IO);
warnings end with `(warning)` and do not change the exit status unless you use `--strict` with `smill check` or `smill build navi`.

## Source and frontmatter {#source}
### Source text {#source-codes}
| Code | Meaning and fix |
| --- | --- |
| `SRC001` | Invalid UTF-8. Re-save the file as UTF-8 without a byte-order mark. |
| `SRC002` | The file exceeds 8 MiB. Split it into several sheets. |

### Frontmatter {.col-span-2 #frontmatter-codes}
| Code | Meaning and fix |
| --- | --- |
| `FM001` | The file must start with a line containing exactly `---`. |
| `FM002` | The closing `---` line is missing. |
| `FM003` | The YAML is invalid or its root is not a mapping. |
| `FM004` | A key appears twice in one mapping. Keep one. |
| `FM005` | Anchors, aliases, merge keys, custom tags, non-finite numbers or nesting beyond 64 levels. Write plain JSON-compatible values. |

### Known fields {#validation-codes}
| Code | Meaning and fix |
| --- | --- |
| `VAL001` | `title` is missing or empty. |
| `VAL002` | `category` or `description` is not a string, or `tags` is not an array of non-empty strings. |
| `VAL003` | `cheatsheet` is not a mapping, `cheatsheet.schema` is not `1`, `cheatsheet.tool` is empty, or `navigation` is not a mapping with only a non-blank string `group` and an integer `order` from 0 to 1000000. |

## Structure and attributes {#structure}
### Sections and cards {#structure-codes}
| Code | Meaning and fix |
| --- | --- |
| `MD001` | A `###` card appears before any `##` section. Add a section heading above it. |
| `MD002` | A section or card heading has no title text. |
| `ID001` | Two headings use the same explicit `#id`, or an explicit ID collides with an earlier generated one. Rename one. |

### Heading attributes {.col-span-2 #attribute-codes}
| Code | Meaning and fix |
| --- | --- |
| `ATTR001` | Unsupported, malformed or misplaced token, for example `{.unknown}`, `{.col-span-0}` or `{.cols-2}` on a card. See the [attribute table](format/gfm-plus.md#heading-attributes). |
| `ATTR002` | The same setting is given twice, for example two `.col-span-` values. |
| `ATTR003` | An attribute group outside a root `##` or `###` heading, such as a `{.marker-none}` line after a list. Remove it, or escape the brace (`\{`) to show it as text. |

## Website export {#export}
### Errors {#render-errors}
| Code | Meaning and fix |
| --- | --- |
| `REN001` | The renderer met content it cannot represent, or the project could not be assembled. Report it with the input. |
| `REN002` | A link uses an unsafe or unsupported URL. Use relative, `http`, `https` or `mailto` links. |
| `REN003` | Two inputs produce the same page route or basename, a route is reserved (`assets`, `brand`, `tags`), two tag labels map to the same tag route, `--theme` or `--palette` names an unknown value, or a base path or GitHub base URL is invalid. |

### Warnings {.col-span-2 #render-warnings}
| Code | Meaning and fix |
| --- | --- |
| `REN101` | Raw HTML is shown as inert text. Use Markdown instead. |
| `REN102` | An image is kept as a link and never fetched. |
| `REN103` | An internal link names no selected sheet, or its fragment matches no section or card ID. Fix the link, add an explicit `{#id}`, or pass `--source-base-url` for links to repository files. |

## navi export {#navi}
### Errors and warnings {.col-span-2 #navi-codes}
| Code | Meaning and fix |
| --- | --- |
| `NAV000` | `smill build navi --strict` found warnings and wrote nothing. Fix them or drop `--strict`. |
| `NAV001` | Unsupported `--shell` profile. Use `bash`, `zsh`, `fish` or `nu`. |
| `NAV006` | Invalid or unknown `cheatsheet.navi` metadata, a reference to a missing/unmarked card or generated ID, a wrong `shell`, conflicting complete old/new parameter declarations, or an unusable tag. |
| `NAV008` | Two exported entries share a pinned navi identity hash; navi would silently drop one. |
| `NAV101`–`NAV106` | Warnings: a marked card was skipped (shape, command text, placeholder, declarations), the license is unstated, or a sheet exported nothing. `NAV007` and `NAV107` are retired: frontmatter is preserved in the `.navi.json` sidecar. See [navi export](format/navi-profile.md). |

## Input and output {#io}
### Files and streams {#io-codes}
| Code | Meaning and fix |
| --- | --- |
| `IO001` | An input cannot be read: missing, not a regular `.md` file, a symlink, or an empty directory selection. |
| `IO002` | Writing to stdout failed, for example a closed pipe. |
| `IO003` | The output exists already or cannot be created. Choose a new path; outputs are never overwritten. |

### Import {#import-codes}
Reference import adds `IMP001`–`IMP003` (errors) and `IMP101`–`IMP111`
(warnings about archived or mapped legacy syntax). See
[Reference import](imports.md) for the full list.

## Checking in CI {#ci}
### Fail on every warning {.col-span-2 #strict}
`--strict` renders your sheets in memory, exactly as `smill build hugo` would, and
turns every warning into an error. Nothing is written.

```sh
smill check --strict sheets
```

### GitHub Actions {.col-span-2 #github-actions}
A minimal job that builds `smill` from a pinned SheetMill revision and checks the
sheets in your repository. Replace the revision with the commit you trust.

```yaml
name: sheets
on: [push, pull_request]
permissions:
  contents: read
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
        with:
          persist-credentials: false
      - uses: actions/setup-go@v6
        with:
          go-version: 1.27.2
      - run: go install github.com/wstein/sheetmill/cmd/smill@<revision>
      - run: smill check --strict sheets
```

Pin third-party actions to full commit SHAs in production workflows.

## Shared command-card diagnostics

| Code | Meaning |
| --- | --- |
| `CMD001` | Invalid shared `cheatsheet.cards` mapping, card settings, parameter name/type/description or unknown key. |
| `CMD002` | A `.command` card lacks an explicit ID, or shared metadata refers to a missing, unmarked or automatically identified card. |

Source diagnostics point to the key/card; direct IR without key positions falls
back to document start. Legacy-only selection/metadata keeps existing NAV behavior.
Conflicting old/new complete declarations use NAV006; NAV008 collision behavior
and pinned hashes are unchanged. See [command cards and migration](format/command-cards.md).
