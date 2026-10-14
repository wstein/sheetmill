# Data model — 014

## Canonical source metadata

`navigation` is optional. When present, it is a mapping with only optional
`group` (string containing non-whitespace) and `order` (integer 0–1,000,000).
`{}` is valid. Null, scalar or array values, unknown subkeys and coercion are invalid.
The existing metadata map retains original values; omitted fields stay omitted.
IR schema remains 1; no new typed field is added to `model.Document`.

## Derived navigation settings

The shared validator returns `Group string` (empty only for omission) and
`Order *int64` (nil for omission, non-nil zero for explicit zero). These are runtime
settings, not synthesized source metadata. Direct IR uses canonical decoded scalar
types; an integral `float64` is not accepted as an integer.

## Presentation group

For each Hugo page, choose explicit group, otherwise original category, otherwise
the existing empty uncategorized key displayed as Other references. Group identity
uses exact strings. Named groups sort by UTF-8 bytes; the fallback sorts last.
An explicitly named Other references group remains a named group, distinct from
the fallback, preserving current category behavior rather than inventing aliases.

## Shelf member ordering

Sort only copied members by `(order presence, order value, lowercase title,
original title bytes, route)`, with present orders first. Lowercasing uses Go's
Unicode `strings.ToLower`, not host locale. Equal orders are permitted without
warnings. Route is already unique. Source titles, tags and categories are retained.
The route-ordered master page list is unchanged so stable related-score ties,
search and tag-catalog order remain unaffected.

## Import provenance

Any root Reference navigation value is removed from active metadata before
canonical validation and stored unchanged under source.original_metadata.navigation.
`IMP102` records inert archival at the original navigation-key span. Navigation is
never generated from other upstream fields.

## Navi companion

Existing companion `metadata` includes the canonical navigation mapping verbatim
as semantic values. Its schema, exported-card list and `.cheat` contents do not
change because of presentation metadata. No reverse import is introduced.
