# Public contract — 014

Implemented under the 2026-10-14 handoff; public format/schema/diagnostic docs
carry the same contract.

```yaml
---
title: SheetMill CLI
category: reference
tags: [sheetmill, smill, markdown]
navigation:
  group: SheetMill
  order: 20
---
```

- Root navigation is optional; when present it must be a mapping. `{}` is valid.
- The only keys are `group` and `order`; a misspelling or deferred `title` fails.
- Group must be a string with non-whitespace content, retained exactly. Omission
  falls back to category; invalid supplied group does not silently fall back.
- Order is a canonical integer 0–1,000,000 inclusive. No string, Boolean, null,
  fractional or integral-looking float coercion. Omission differs from explicit 0.
- Invalid navigation uses VAL003, positioned at its original key where available;
  direct IR uses 1:1 document start. The website renderer applies the same focused
  rule as source validation and returns no partial project on failure.
- Ordered sheets precede unordered ones and compare numerically ascending; ties
  compare Unicode-lowercased title,
  original UTF-8 title bytes, then route. Duplicate orders are silent.
- Only library shelf placement changes. Category/tag metadata, related relevance,
  source/card order, URLs, command text and identities retain existing semantics.
- Incoming Reference navigation of any shape is archived as inert provenance
  before validation, with IMP102. Independently authored navigation survives in
  metadata-preserving exports and the existing navi companion.

## Compatibility

IR schema 1 and omitted fields are unchanged. This does not guarantee complete
source backward compatibility: navigation was formerly generic JSON-safe metadata
and is now a reserved closed namespace. Existing navigation objects containing
other keys must be renamed to an author-owned metadata key or brought into the
new contract. Other unknown root metadata remains extensible/preserved.
The JSON Schema describes metadata after YAML decoding; strict YAML scalar-type
rules additionally reject `order: 1.0`, even though JSON Schema's mathematical
integer semantics cannot by themselves distinguish that spelling from 1.

No group-order file, manifest override, navigation title, stable sheet ID,
publication change or deployment permission is part of this interface.
