# Research — 014 library navigation

Baseline: main `de3e08e`, inspected 2026-10-14 in an isolated worktree. Static
inspection only; future navigation behavior has not been implemented or tested.
A separately assigned Sol researcher independently verified the claims below.

## Verified claims

| Claim | Evidence and consequence |
| --- | --- |
| Metadata diagnostics are fatal | `internal/frontmatter/frontmatter.go`, `Parse`, calls `validate.Metadata` and returns an empty result on diagnostics; `internal/markdown/markdown.go`, `Parse`, propagates them. Unknown-key warnings cannot be inserted into this return channel. |
| Import already validates | `internal/importer/reference/reference.go`, `Convert`, calls `mapLegacyMetadata`, then `validate.Metadata`; existing visual archival is later. Navigation must be archived before validation. |
| Direct Hugo IR bypasses source validation | `internal/render/hugo/hugo.go`, `Render`, currently checks document presence/schema, not all metadata. Call the focused shared navigation helper there rather than introducing unrelated full metadata validation. |
| Sidecar preservation exists | `internal/render/navi/navi.go`, `Render`, marshals `s.Document.Metadata` into the schema-1 companion. No new sidecar field/model is needed. |
| Shelf data feeds both views | `shelves()` in `internal/render/hugo/hugo.go`; `theme/layouts/home.html` (both layouts) and `theme/layouts/partials/library.html` consume its data. |
| Related ties depend on route order | `related()` uses stable score sorting over route-ordered `pages`. Sort copied shelf-member slices only; do not reorder that main slice. |
| Generic IR already preserves metadata | `internal/model/model.go`, `Document.Metadata`, is a JSON-safe map; parser `TestGolden` discovers `.md` and compares `.md.json` under `testdata/reference-compatible/`. No typed IR field/version change is necessary. |
| Nested source spans exist | `frontmatter.Decode` records dotted key locations and returns YAML integers as `int64`. Use these positions for source errors. |

## Decisions

### Shared validation, two consumers

Decision: add a focused `validate.Navigation` rule with a derived
`NavigationSettings` result (`Group string`, `Order *int64`) and diagnostics.
`validate.Metadata` delegates to it; Hugo calls it after nil/schema checks using
its document-start fallback span. Both paths reject the same cases.
Rationale: one rule source without silently trusting hand-built IR or broadening
Hugo's direct-IR checks for unrelated metadata. Derived defaults never enter source
metadata. Alternative rejected: duplicating checks or revalidating every core field
inside Hugo as an incidental change.

### Numeric and group semantics

Decision: accept canonical `int64` order only, inclusive 0–1,000,000. Omission is
nil, explicit zero is a present pointer. A source float `1.0` and direct-IR
`float64(1)` are rejected; no coercion or new JSON IR loader is introduced.
Group is exact/case-sensitive, validated for non-whitespace content without
rewriting the original value. Empty mapping is valid; explicit null is not.
Rationale: match existing decoder types and preserve authored metadata.

### Ordering boundaries

Decision: compare shelf members by present order first, numerical order, Unicode
`strings.ToLower(title)`, original UTF-8 bytes, then route. Duplicate orders are
normal and silent. Group names retain the existing byte-wise ordering, with the
empty uncategorized key last. Only copies inside `shelves()` are sorted.
Rationale: home/sidebar consistency without changing `related()`, tag results,
search/library route order, categories or URLs. Alternative rejected: globally
sorting `pages`, which would change unrelated stable relevance ties.

### Import archival and diagnostics

Decision: archive any incoming root Reference navigation before `validate.Metadata`,
under `source.original_metadata.navigation`, with existing `IMP102` and the original
key span. Preserve the whole value even when invalid under the canonical namespace.
No inference from category/title/path/tool. A small single-key archival helper may
be shared with the current late visual loop; do not move that whole loop or change
existing unrelated finding order.
Rationale: upstream metadata must not activate presentation accidentally or become
an import failure merely because a new canonical namespace is reserved.
Alternative rejected: preserving it as active canonical placement or validating
before archival. `VAL003` is reused and its public meaning explicitly broadened.

### Compatibility and exclusions

Decision: keep IR schema 1 and metadata omission unchanged; reserve only the new
navigation namespace as closed. Root metadata remains extensible. Existing generic
navigation can now fail canonical source validation: this is a deliberate source
compatibility restriction, not a promise of complete backward compatibility.
The source JSON Schema describes post-decoding metadata; JSON Schema's integer
semantics alone do not encode the YAML distinction between `1` and `1.0`. Keep
that strict source rule explicit in the format docs and parser tests.
Group-order config, manifests, navigation title, stable IDs, content reordering,
execution and deployment are deferred. No new dependencies or renderer file I/O.
