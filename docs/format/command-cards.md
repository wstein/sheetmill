---
title: "Shared command cards (proposal)"
description: "Target-neutral command cards, shared parameters and identity-preserving migration."
category: reference
---
**Proposed, not implemented.** Feature [015](../../specs/015-command-cards/spec.md)
extends SheetMill's core source contract; this documentation PR does not add
parser or exporter support. Use the [current navi syntax](navi-profile.md) today.
The future contract below takes effect only with a separately authorized implementation.

## Author once

Command semantics belong to the source document. Parameters live under
`cheatsheet.cards.<id>.params`, without a `navi` container:

````markdown
---
title: SheetMill CLI
tags: [sheetmill, validation]
cheatsheet:
  tool: smill
  cards:
    check-sheet:
      params:
        sheet:
          type: path
          description: Markdown sheet to validate
---

## Validation

### Check a sheet {#check-sheet .command}

Validate a sheet before publishing it.

```bash
smill check "<sheet>"
```
````

`cheatsheet.schema: 1` remains optional. `.command` declares semantics, not an
execution action or a layout style. A shell example without a command marker is
not automatically exported. No new visual controls or renderer behavior beyond
navi consumption is promised by this proposal.

## Shared declarations

| Location | Proposed rule |
| --- | --- |
| `cheatsheet.cards` | Optional mapping keyed by explicit card ID; empty mapping allowed |
| `.cards.<id>` | Mapping with only optional `params` |
| `.params` | Mapping, empty or omitted for no declared parameters |
| `.params.<name>` | Mapping with required `type`, optional `description`, no other fields |
| `type` | Exactly `string`, `int` or `path` |
| `description` | Nonempty single-line string under the current parameter-description rules |

Descriptions retain the current restrictions: no leading/trailing whitespace,
control or format characters, or U+2028/U+2029 separators; descriptions are not
silently normalized.

Null, scalar or sequence mappings and coercion fail. Parameter names match
`[a-z][a-z0-9_]{0,31}`. Card IDs keep the existing ASCII letter followed by letters,
digits, `_` or `-` grammar. Parameters stay card-local: two cards may declare the
same name differently. A zero-parameter command needs no frontmatter card entry.

Core checking validates shared declarations and their references to explicit
command-card IDs. `.command` cards require explicit IDs. Direct-IR renderer callers
apply the same shared rules. Proposed errors are CMD001 (shared metadata) and
CMD002 (explicit ID/card reference), with source key/card positions or 1:1 for
positionless IR. These codes are reserved here, not implemented diagnostics.

Navi still applies its own one-fence shape, supported fence languages, raw payload,
placeholder-use, quote and output-profile rules. Parameter hints neither enforce
runtime types nor quote substituted input. Nothing is executed by SheetMill.
Defaults, choices, providers, previews and shared contexts remain out of scope.

## Migration and compatibility

After implementation:

1. Keep the filename, path/root selection, explicit card ID, title, command bytes,
   tool/tags, license, attribution and build options unchanged.
2. Replace the heading marker `.navi` with `.command`.
3. Move each complete card declaration from `cheatsheet.navi.cards.<id>` to
   `cheatsheet.cards.<id>`. Do not rename parameters or reformat the command.
4. Build with the same `--shell` profile. New syntax needs no navi frontmatter.
   Fence language describes command dialect; the build option selects the output
   profile. If removing legacy `cheatsheet.navi.shell`, explicitly retain that
   profile in the build invocation/configuration.
5. Verify byte-identical `.cheat` output and unchanged per-entry pinned identities
   before adopting the migrated source. No automated migration tool is included.

| Input after implementation | Behavior |
| --- | --- |
| Legacy `.navi` plus `cheatsheet.navi.cards` | Supported with existing export behavior |
| New `.command` plus `cheatsheet.cards` | Shared semantics consumed by navi |
| Either marker with either declaration location | Compatibility alias resolves to the same card |
| Both markers on one card | Export once, with the same identity |
| Old/new declarations for disjoint card IDs | Combine by card ID |
| Old/new declarations for the same card | Coalesce only if complete normalized declarations agree; otherwise NAV006, no files |
| Invalid legacy metadata alongside valid new metadata | Still fails; new syntax does not mask old errors |

Omitted `params` and `params: {}` mean the same empty declaration set. Otherwise
compare complete semantic per-card declarations, including types and descriptions,
independently of map key order. Do not merge subsets of one card's parameter maps.
Both present forms are validated before resolving conflicts.

Legacy `cheatsheet.navi.shell` stays an optional compatibility guard when present;
a mismatched profile remains NAV006. Removing it does not promise identical
cross-profile diagnostics: a wrong-language new command follows the existing
navi warning-and-skip rules. Output parity concerns successful equivalent exports
under the same profile, not failures from deliberately different options.
Legacy-only missing-ID/metadata diagnostics retain existing NAV behavior; core
CMD rules apply to new `.command` cards/shared declarations. In a mixed document,
a shared declaration may reference an explicitly identified `.navi` alias card.

No default deprecation warning is added, preserving `--strict` behavior for valid
legacy sheets. No alias-removal date is set. Removing compatibility requires a
separate versioned decision; this proposal does not rewrite documents automatically.

## Identity and byte guarantees

Migration must not change the entire `.cheat` byte stream, including comment
headers, format version, context/tag lines, normalized titles, command payloads,
entry order, encoding and final newlines, for otherwise equivalent successful inputs.

The first tag stays `sheetmill:<encoded relative path>#<encoded explicit id>!`.
Navi's pinned item hash stays computed from the existing full tag-line text plus
normalized card title with the exact current algorithm. Neither the marker nor
the metadata location enters identity. Preserve tool/tag ordering and context
encoding; do not introduce a new hash scheme or reset saved identities.
NAV008 continues rejecting batch collisions before anything is written.

The future implementation must compare complete old/new `.cheat` bytes **and**
fixed expected hashes for every entry. Merely comparing old and new hashes is
insufficient: both could drift together. Existing committed hash vectors remain.
Dual-marker/mixed fixtures must prove no extra entry appears.

Companions preserve complete authored semantic metadata, so old/new `.navi.json`
bytes may differ. Their exported-card IDs/titles/contexts must agree. Do not rewrite
the source metadata to aliases merely to make companions equal. IR schema remains 1.

## Source compatibility and imports

`cheatsheet.cards` was previously generic metadata. Reserving it introduces a
source compatibility restriction even with unchanged IR version: conflicting
custom fields must be renamed or brought into the new contract. Other unknown
metadata remains preserved. The new `.command` marker currently fails parser
validation; it is not usable until implementation.

Import adapters must archive upstream `cheatsheet.cards` and `.command` markers
as inert provenance before core validation/selection, never infer or activate
command behavior. Preserve original values/marker text and unrelated metadata;
legacy import behavior is not otherwise changed by this proposal.

See [015's plan](../../specs/015-command-cards/plan.md) and
[pending tasks](../../specs/015-command-cards/tasks.md) for future verification.
