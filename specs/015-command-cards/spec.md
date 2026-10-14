# Feature Specification: Target-neutral command cards

**Feature Branch**: `codex/015-command-cards`
**Created**: 2026-10-14
**Status**: Implemented under the 2026-10-14 handoff; code re-review clear; final-head evidence recorded on the PR
**Input**: Shared `.command` semantics and `cheatsheet.cards.<id>.params`, with
migration from 006 and byte-identical navi output and pinned entry identities.

## User Scenarios & Testing

### User Story 1 — Author once for multiple targets (Priority: P1)

An author describes a command and its parameters without a navi-specific metadata
container. Renderers consume the same semantic description.

**Why this priority**: Parameters describe source content, not one output format.
**Independent Test**: Check and inspect a shared command card; confirm preserved
metadata and explicit identity, with no command execution or output-specific setup.

**Acceptance Scenarios**:
1. **Given** a card `{#check-sheet .command}` and matching
   `cheatsheet.cards.check-sheet.params`, **When** checking it, **Then** core rules
   validate declarations and references without requiring `cheatsheet.navi`.
2. **Given** ordinary fenced shell examples, **When** exporting, **Then** they do
   not become command entries automatically.
3. **Given** a shared declaration with an unknown field or missing/unmarked card,
   **When** checking or rendering directly supplied IR, **Then** positioned errors
   reject it consistently rather than silently dropping metadata.

### User Story 2 — Migrate without losing navi identities (Priority: P1)

An author changes only the marker and parameter-metadata location while retaining
the same card, path, content and build profile.

**Why this priority**: Existing navi identities must survive source modernization.
**Independent Test**: Compile paired legacy/new documents and compare the complete
`.cheat` bytes and every exported entry's pinned identity hash against fixed baselines.

**Acceptance Scenarios**:
1. **Given** equivalent valid legacy/new inputs and identical export options,
   **When** exporting, **Then** `.cheat` output is byte-identical, including headers,
   contexts, tag order, normalized titles, entry order and payloads.
2. **Given** either marker, or both on one card, **When** exporting, **Then** each
   selected card is emitted once and NAV008 collision checks remain unchanged.
3. **Given** old/new parameter declarations for one card, **When** both are equal,
   **Then** they coalesce; differing declarations fail rather than taking precedence.

### User Story 3 — Adopt incrementally (Priority: P2)

An author retains existing valid documents and migrates cards at their own pace.

**Why this priority**: A core source revision must not invalidate working exports.
**Independent Test**: Exercise legacy-only, new-only and mixed documents, including
strict builds and metadata companions, without running navi or a shell.

**Acceptance Scenarios**:
1. **Given** valid legacy input, **When** upgrading, **Then** existing export
   behavior remains available with no new default deprecation warning or strict failure.
2. **Given** semantically equivalent syntax variants, **When** reading companions,
   **Then** exported-card IDs/titles/contexts agree while full authored metadata
   remains preserved; companion bytes need not be equal.

### Edge Cases

Empty shared cards/params mappings are valid; null/list/scalar mappings are not.
A command with no parameters needs no frontmatter card entry. Shared declarations
must reference explicit command-card IDs; `.navi` is a compatibility alias for
selection and reference resolution. Both markers select once. Per-card settings
must agree as complete semantic mappings when duplicated across namespaces;
missing `params` and `params: {}` normalize to the same empty declaration set.
Disjoint card IDs combine; same-card parameter maps are never partially merged.
Malformed legacy metadata is not made valid by an accompanying new declaration.

## Requirements

### Functional Requirements

- **FR-001**: Introduce `.command` on card headings and optional
  `cheatsheet.cards.<id>.params`; source format docs MUST label it unimplemented
  until a separately authorized implementation ships. No new navi container is required.
- **FR-002**: Shared card settings allow only optional `params`. Parameter names
  match `[a-z][a-z0-9_]{0,31}`. Each parameter requires `type` in `string|int|path`
  and permits only optional nonempty single-line `description` under existing
  description rules (no leading/trailing whitespace, controls, format characters
  or U+2028/U+2029); unknown keys fail.
  IDs retain the existing explicit ASCII ID grammar. No implicit type coercion.
- **FR-003**: Core checking validates shared mapping shapes/types and explicit
  command-card references. Shared rules also protect direct-IR consumers.
  CMD001 covers shared metadata; CMD002 covers missing explicit IDs and dangling/
  unmarked references. Legacy-only errors retain their current NAV codes/behavior.
- **FR-004**: Treat `.navi` as a compatibility selection alias; retain legacy
  `cheatsheet.navi.cards` and its shell guard during migration. Equal duplicate
  declarations coalesce; conflicts are fatal NAV006 before writing. No default
  deprecation warnings, alias-removal date or automatic source rewriting.
- **FR-005**: Fence language remains command dialect; `smill build navi --shell`
  selects the output profile. Existing fence/profile matching, structural/payload/
  placeholder checks, warning-and-skip and `--strict` behavior remain navi rules.
- **FR-006**: Equivalent valid syntax variants MUST emit byte-identical `.cheat`
  files and identical pinned per-entry hashes. Preserve context encoding and first
  tag, tool/tag ordering, title normalization, payload, entry ordering, header
  format/version, hash algorithm/input and NAV008 batch collision detection.
- **FR-007**: Keep schema-1 IR and existing metadata/heading fidelity. Derive shared
  semantics without rewriting authored metadata or injecting defaults. Companions
  preserve original semantic metadata and the same exported-card identities.
- **FR-008**: Reserving `cheatsheet.cards` restricts previously generic metadata;
  document compatibility and diagnose malformed values. Unknown unrelated metadata
  remains preserved. Import adapters MUST NOT accidentally activate upstream shared
  declarations/`.command` markers; archive those inputs inertly before core checks.
- **FR-009**: No navi, shell, source snippet, provider, preview command or linked
  resource executes. Type hints do not enforce runtime types, quote substitutions
  or make free input safe. No new rendering UI or execution capability is implied.

### Key Entities

- **Command card**: explicitly identified semantic card, with literal code content.
- **Card parameter declaration**: card-local type/description hints shared by targets.
- **Exported identity**: pinned navi hash derived from the existing emitted tag line
  and normalized title; source marker/metadata location are excluded.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Every paired supported migration fixture yields zero `.cheat` byte
  differences and identical fixed expected per-entry hashes and contexts.
- **SC-002**: Legacy-only fixtures keep existing outputs and strict-build behavior;
  dual aliases never duplicate entries and conflicting declarations always fail atomically.
- **SC-003**: Shared invalid inputs fail source and direct-IR checks consistently;
  no existing parser/import golden changes without a separately justified semantic change.
- **SC-004**: Companions retain all authored metadata and unchanged exported-card
  identities; imported shared fields/markers remain inert provenance.

## Assumptions

The user authorized T001–T011 implementation, subagents, Sol review and a PR on
2026-10-14, recorded in AGENTS.md. Automated migration tools, navi execution,
PDF, deployment and merge remain excluded. 015 owns core source semantics; 006
consumes them for navi. Existing valid legacy exports remain compatible.
For parity, source path/root resolution, explicit IDs, titles, tags, tool, license,
attribution, command bytes and build options remain identical. Legacy shell guards
continue to apply when present; parity covers successful matched-profile exports,
not identical diagnostics across different profiles or malformed syntax.
