---
title: "Proposed navi authoring profile"
description: "Explicit cards and free-input parameters for a future navi exporter."
category: reference
---
This is the proposed `006` format contract, **not an implemented export feature**.
Today `.navi` is reserved and nested frontmatter is preserved generically; neither
`smill check` nor the JSON Schema validates these proposed navi rules. No navi CLI
command is available. See the [006 specification](https://github.com/wstein/sheetmill/blob/main/specs/006-navi-renderer/spec.md)
and [source evidence and pending checks](https://github.com/wstein/sheetmill/blob/main/specs/006-navi-renderer/research.md).

## One explicit command card

Keep ordinary GFM prose, tables and layout. Only a card with an explicit ID,
`.navi` and matching metadata is proposed for export. Its sole code block must be
a root-level `bash` fence containing one physical command line. Multiple blocks,
indented/nested blocks, other languages, multiline scripts and heredocs fail
selection rather than exporting a guessed command. Unmarked cards are ignored.

````markdown
---
title: SheetMill CLI
tags: [sheetmill, validation]
cheatsheet:
  schema: 1
  tool: smill
  navi:
    profile: bash-free-input-v1
    cards:
      check-sheet:
        params:
          sheet:
            type: path
            description: Markdown sheet to validate
---

## Validation

### Check a sheet {#check-sheet .navi}

Validate a sheet before publishing it.

```bash
smill check "<sheet>"
```
````

`bash-free-input-v1` targets navi v2.24.0 and Bash 5.2 syntax. It does not set
navi's shell; consumers must configure compatible Bash themselves. Other shell
profiles are deferred. Runtime compatibility tests remain pending; the profile
must pass the specification's non-executing checks before being called tested.

## Proposed metadata schema

| Location | Required value |
| --- | --- |
| `cheatsheet.navi` | Mapping with exactly `profile` and `cards` |
| `.profile` | String, exactly `bash-free-input-v1` |
| `.cards` | Nonempty mapping keyed by explicit `.navi` card IDs |
| `.cards.<id>` | Mapping with exactly `params` |
| `.cards.<id>.params` | Mapping keyed by parameter names; `{}` for no parameters |
| `.params.<name>` | Mapping with exactly `type` and `description` |
| `.type` | String, `string` or `path`; an authoring hint only |
| `.description` | Nonempty string; preserved in sidecar, not a navi prompt guarantee |

Parameter names match `[a-z][a-z0-9_]{0,31}`. Each placeholder is a complete
unescaped double-quoted argument, such as `"<sheet>"`; every declaration is used
and every occurrence declared. Names are local to a card, and repeat occurrences
are permitted. Metadata references to missing/unmarked cards fail. All marked
cards require metadata, including cards with no parameters.

Defaults, enums, suggestion commands, previews, maps, multi-selection and shared
contexts are deferred and diagnosed; no `$` providers or `@` inheritance are
emitted. Unknown navi keys fail. Metadata outside this namespace is preserved
unchanged in the sidecar, not silently discarded.

Navi recognizes more token spellings than this profile: unsupported names,
escaped tokens and placeholders in larger words or other shell contexts fail.
There is no literal-placeholder escape in v1. Angle brackets used as ordinary
shell redirections, for example `cat < input.txt`, are not placeholders. Validate
shell syntax separately instead of balancing all `<` and `>` characters.

## Output mapping and identity

For an input identified relative to the explicit input root as `sheets/smill.md`,
the command entry would be the following illustrative text, not a generated file:

```text
% sheetmill:sheets/smill.md#check-sheet!, sheetmill, validation
# Check a sheet
smill check "<sheet>"
```

Each card gets a unique context: `sheetmill:<encoded-relative-path>#<encoded-id>!`.
Encode UTF-8 bytes outside ASCII letters, digits, `_`, `-`, `.`, `/` as uppercase
`%HH`, including literal `%`, comma, `#` and `!`; reject absolute/parent paths.
The final `!` is a structural terminator, never part of an encoded source value. Keep
the `.md` extension. Paths are relative to the proposed required `--root` directory, use `/` and
retain case; duplicate input identities and output paths fail before writing.
Append root `tags` in source order after the context, removing exact duplicates.
Tags must be trimmed, nonempty, single-line, comma-free and contain none of the
forbidden control/format/separator characters listed in S3. A card title must
be nonempty, trimmed, single-line and obey the same character exclusions; commas
in titles are allowed. These display restrictions fail with `NAV006`, not rewriting.

The terminated context avoids ambiguous concatenation of tags and titles. Before
writing, compute the pinned navi item identity hash for every intended entry
across the export batch and reject duplicate hashes (`NAV008`), including actual
hash collisions. Distinct context strings alone are insufficient: navi hashes
trimmed tags followed immediately by the trimmed description. This protects the
export batch; unrelated cheats loaded later by the consumer remain outside its
validation scope. Titles become descriptions; prose, sections and layout remain
presentation content. No metadata is inferred from prose. Shell payload bytes
are retained, with a terminal LF added only if absent. A payload beginning with
navi markers is rejected, not indented or hidden in fences.

A full future file also starts with `;` provenance metacomments. Header values are
single-line JSON strings; these comments do not create command entries. The
versioned `.navi.json` sidecar retains complete IR metadata and selected card IDs,
parameter hints and profile. Output names retain relative directories, with the
`.md` suffix replaced by `.cheat` and `.navi.json` respectively.

## Free input is trusted shell text

Navi substitutes raw text, including values obtained from its environment. The
double quotes in `"<sheet>"` help ordinary paths with spaces, but do not safely
escape arbitrary quotes, substitutions or shell operators in a value. Neither
`type: path` nor `type: string` enforces runtime validation. Do not claim arbitrary
input remains one safe argument. A leading dash may become a program option;
use a tool-supported `--` or a suitable `./` path where appropriate.

The generated header and eventual user documentation must carry this boundary.
SheetMill only generates text and never runs examples. Choosing to execute the
exported template in navi remains a separate user action. Browser previews keep
all commands inert. Executable providers and arbitrary-input-safe argument
handling require separate design; they are not implied by this profile.
