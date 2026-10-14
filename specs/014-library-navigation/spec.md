# Feature Specification: Library navigation

**Feature Branch**: `codex/014-navigation-complete`

**Created**: 2026-10-14

**Status**: Implemented first slice; local verification complete, PR review pending

**Input**: User-settled optional `navigation.group` and `navigation.order`, closed
namespace, deterministic ordering, inert Reference import and navi preservation.

## User Scenarios & Testing

### User Story 1 - Place related references together (Priority: P1)

An author places SheetMill references in one sidebar group without changing their
subject categories, tags, titles, filenames or links, and chooses their reading order.

**Why this priority**: Product references currently split across subject groups.

**Independent Test**: Build a small library with placement overrides, explicit
orders and unordered sheets; compare the group/member sequence on the home page
and sidebar, including shuffled selections.

**Acceptance Scenarios**:

1. **Given** a reference with `category: reference` and `navigation.group: SheetMill`,
   **When** browsing the library, **Then** it appears in SheetMill while its category
   and related-reference relevance remain unchanged.
2. **Given** orders 20, 0, 20 and an omitted order, **When** browsing their group,
   **Then** zero comes first, tied 20s follow the title/route rules, and the unordered
   sheet follows all explicitly ordered sheets, without duplicate-order warnings.
3. **Given** no navigation group, **When** browsing, **Then** category supplies the
   group; absent/empty category uses the existing Other references fallback.
4. **Given** any permutation of the same inputs, **When** building, **Then** the
   home and sidebar show the same deterministic sequence in both layouts.

### User Story 2 - Catch placement mistakes without hidden defaults (Priority: P1)

An author receives positioned errors for malformed navigation and can distinguish
an explicit priority of zero from an omitted priority.

**Why this priority**: Silent misspellings make navigation unpredictable.

**Independent Test**: Check boundary and malformed inputs and compare preserved
metadata from valid inputs with existing snapshots.

**Acceptance Scenarios**:

1. **Given** `navigation: {ordr: 10}`, **When** checking/building, **Then** `VAL003`
   reports the unknown key; no partial export is written.
2. **Given** order 0 or 1,000,000, **When** checking, **Then** it is accepted;
   negative, over-cap, floating-point, Boolean, string and null values are rejected.
3. **Given** no navigation or an empty navigation mapping, **When** inspecting,
   **Then** no omitted group/order is materialized and the IR version remains 1.
4. **Given** malformed navigation in directly supplied semantic input, **When**
   exporting the website, **Then** the same navigation rules reject it rather than
   silently ignoring it.

### User Story 3 - Preserve import and export boundaries (Priority: P2)

An author imports upstream content without activating upstream placement metadata,
while independently authored navigation survives metadata-preserving exports.

**Why this priority**: Reserving a previously generic metadata name affects imports.

**Independent Test**: Convert an upstream file with navigation of an invalid
canonical shape, and compare navi exports with/without valid authored navigation.

**Acceptance Scenarios**:

1. **Given** Reference input containing any root navigation value, **When**
   converting, **Then** the complete value is archived in
   `source.original_metadata.navigation`, an import finding identifies it, and no
   active root navigation is generated or retained.
2. **Given** valid authored navigation added to a sheet, **When** exporting navi
   text and its companion, **Then** `.cheat` bytes and exported-card identities
   remain unchanged, and the companion preserves navigation as metadata.

### Edge Cases

- Empty mapping is valid; null/list/scalar navigation is invalid.
- A present group must be a string containing a non-whitespace character; an
  empty/whitespace-only group is invalid rather than requesting category fallback.
- Unknown subkeys, including deferred `title`, are errors. Unknown metadata
  outside the newly closed namespace remains preserved under existing rules.
- Integer zero is explicit; integral-looking float `1.0` is still invalid source.
- Duplicate orders, duplicate titles, Unicode case variants and input permutations
  have deterministic tie-breaks and produce no duplicate-order warnings.
- Reference navigation is archived before canonical navigation validation, even
  when its value would otherwise be rejected. Existing unrelated import finding
  order remains unchanged.
- Group labels are exact and case-sensitive; no implicit trimming or case merging
  changes authored identity. Trimmed emptiness is used only for validation.
- The uncategorized fallback retains its existing distinct empty internal key and
  sorts last; an explicitly named Other references group is an ordinary named group.

## Requirements

### Functional Requirements

- **FR-001**: Support an optional root `navigation` mapping with only optional
  `group` and `order`; no defaults are written into metadata.
- **FR-002**: A present group MUST be a non-whitespace string. A present order MUST
  be an integer in inclusive range 0–1,000,000; nulls and coercion are disallowed.
- **FR-003**: Invalid navigation, including unknown subkeys, MUST fail with
  `VAL003` using the source key location when available, otherwise document start.
- **FR-004**: Effective group MUST use an explicit navigation group, otherwise
  category, otherwise the existing Other references fallback. Group order remains
  existing byte-wise name order with the uncategorized fallback last.
- **FR-005**: Within each group, explicit orders MUST precede omitted orders and
  compare numerically ascending. Ties MUST compare Unicode-lowercased page title,
  original UTF-8 title bytes, then route. Duplicate orders MUST be silently allowed.
- **FR-006**: Home and sidebar MUST consume the same ordered group membership.
  Titles, categories, tags, routes, related scoring and related tie-breaks MUST not
  change as a consequence of placement metadata.
- **FR-007**: Validation rules MUST be shared across source checking and direct
  website-rendering input; direct callers cannot bypass navigation validation.
- **FR-008**: Reference navigation MUST be archived inertly before canonical
  validation, with `IMP102`; the importer MUST NOT infer navigation from other data.
- **FR-009**: Existing metadata-preserving IR and navi companions MUST retain
  authored navigation. No sidecar import or new round-trip feature is introduced.
- **FR-010**: Document that reserving navigation can reject previously accepted
  generic metadata despite unchanged IR version 1, and widen `VAL003`'s meaning.
- **FR-011**: Verification MUST keep existing parser/import golden bytes unchanged,
  add a small navigation metadata golden, and separately test multi-sheet website
  ordering, permutation determinism, unchanged related relevance and navi identities.

### Key Entities

- **Navigation metadata**: optional placement group and optional bounded order.
- **Effective library group**: a presentation grouping with ordered sheet members;
  independent of the sheet's subject category and tags.
- **Import provenance**: preserves upstream navigation without activating it.

## Success Criteria

### Measurable Outcomes

- **SC-001**: All 24 permutations of a four-sheet ordering case, plus seeded shuffles of
  the full fixture, yield byte-identical exported files and identical home/sidebar
  sequences.
- **SC-002**: Every invalid navigation case is rejected with the documented code
  and position; boundary integers and absent/zero distinctions behave as specified.
- **SC-003**: All existing parser/import snapshots remain byte-identical, and
  baseline references retain their URLs and related-reference results.
- **SC-004**: Upstream navigation is preserved only as inert provenance; authored
  navigation survives the companion with zero changes to command text or card identities.
- **SC-005**: Home and sidebar ordering agrees in both layouts, at root/subpath
  mounts and with JavaScript disabled, without changing the existing navigation controls.

## Assumptions

- The user settled the design and authorized implementation and a PR on
  2026-10-14, recorded in AGENTS.md. Merge and deployment remain separately gated.
- Only per-sheet group/order ships in the first slice. Group-order configuration,
  manifest overrides, navigation title, stable sheet IDs, library content reordering,
  PDF, navi execution, deployment and merge are excluded.
- Public format/schema/diagnostic docs describe the implemented contract.
