# Feature Specification: Sheet freshness

**Feature Branch**: `codex/017-freshness`

**Created**: 2026-10-16

**Status**: Specification under independent review; no implementation yet

**Input**: User request (2026-10-16), prompted by a comparison with a public
cheatsheet site: show an `updated` date and the documented tool version on library
shelf cards and on the sheet header, deterministically.

## User Scenarios & Testing

### User Story 1 - Readers see how fresh a sheet is (Priority: P1)

A reader scanning the library or opening a sheet sees which tool version the sheet
documents and when the author last updated it, without JavaScript and without the
page ever depending on the build date.

**Why this priority**: It is the whole purpose of the feature.

**Independent Test**: Build a small library with/without the fields in all three
themes; assert the markup of home cards, index rows and sheet headers, and that two
builds with different Hugo `--clock` values are byte-identical.

**Acceptance Scenarios**:

1. **Given** a sheet with `cheatsheet.tool_version: "2.56.0"` and
   `cheatsheet.updated: 2026-10-14`, **When** the editorial home page is built,
   **Then** its shelf card shows `Version 2.56.0` and
   `Updated <time datetime="2026-10-14">2026-10-14</time>` in one compact line.
2. **Given** the same sheet, **When** the sheet page is built in `editorial`,
   `classic-compact` and `reference-cards`, **Then** the header shows both facts in
   the same wording and the compact layouts' index rows show them as short text.
3. **Given** a sheet with neither field, **When** building, **Then** no freshness
   element, label, separator or empty wrapper is emitted and nothing is reported.
4. **Given** only one of the two fields, **When** building, **Then** only that fact
   is rendered, without a dangling separator.
5. **Given** the same sources built with `--clock 2000-01-01T00:00:00Z` and with
   `--clock 2031-06-01T00:00:00Z`, **When** comparing the output, **Then** every file
   is byte-identical.

### User Story 2 - Authors get strict, positioned errors (Priority: P1)

An author who mistypes a date or writes an unusable version gets a source-positioned
error instead of a silently odd page.

**Independent Test**: Table tests over valid and hostile values through the source
decoder and through direct IR input.

**Acceptance Scenarios**:

1. **Given** `updated: 2026-02-30`, `2026-1-5`, `20261016`, `2026-10-16T09:00:00Z`,
   `"16 Oct 2026"`, a Boolean, null, list or number, **When** checking, **Then**
   `VAL004` points at the `cheatsheet.updated` key.
2. **Given** `updated: 2024-02-29` (leap day) and `1970-01-01` and `9999-12-31`,
   **When** checking, **Then** they are accepted.
3. **Given** `tool_version` that is empty, blank, longer than 64 code points,
   multi-line, has surrounding whitespace, a control, bidi or other invisible format
   character, or is not a string, **When** checking, **Then** `VAL004` points at it.
4. **Given** `updated: 2999-01-01`, **When** checking, **Then** it is accepted: the
   compiler has no clock (constitution 1), so "future" is not decidable.
5. **Given** malformed freshness in directly supplied semantic input, **When**
   exporting the website, **Then** the same rules reject it and no project is returned.

### User Story 3 - Imports and exports keep their boundaries (Priority: P2)

Upstream Reference content never activates or fails on freshness metadata, and navi
exports keep their command bytes.

**Acceptance Scenarios**:

1. **Given** Reference input with `cheatsheet.updated` or `cheatsheet.tool_version`
   of any shape, **When** converting, **Then** the complete values are archived in
   `source.original_metadata.cheatsheet`, an `IMP112` warning identifies each, and
   no active freshness metadata is generated; a date is never inferred.
2. **Given** a sheet with valid freshness, **When** running `smill build navi`,
   **Then** every `.cheat` file is byte-identical to the build without the fields,
   exported-card identities are unchanged, and the metadata companion carries the
   fields verbatim as ordinary metadata.
3. **Given** the same library, **When** reordering input files or changing dates,
   **Then** shelf order, related results and tag pages do not change.

### Edge Cases

- YAML `updated: 2026-10-14` (a timestamp scalar) and `"2026-10-14"` (a string) are
  both preserved as the same string and both valid; only date-only ISO 8601 is valid.
- Only ASCII digits are accepted; Unicode digits, `+`/`-` year signs and
  week/ordinal ISO forms are errors.
- `cheatsheet` present but not a mapping keeps its existing `VAL003`; freshness is
  not additionally reported.
- Unknown `cheatsheet.*` keys remain preserved generic metadata (no closed namespace).
- Hostile text (`<script>`, quotes, `&`) in `tool_version` is HTML-escaped in text;
  `datetime` is only ever a validated digits-and-hyphens value.
- The version is shown as authored; no normalization, `v` prefix or semver parsing.
- A sheet whose `cheatsheet.tool` is absent still shows its version.
- A date later than the build is shown as authored.

## Requirements

### Functional Requirements

- **FR-001**: Support optional `cheatsheet.updated` and `cheatsheet.tool_version`;
  no defaults are written into metadata or IR; `cheatsheet.tool` is unchanged and
  is reused as the existing tool name. IR schema stays 1.
- **FR-002**: `cheatsheet.updated` MUST be a string spelling a real proleptic
  Gregorian calendar date exactly as `YYYY-MM-DD` (ASCII digits), year 1970–9999.
  YAML timestamps arrive as strings and are accepted; any time, zone, other
  spelling or non-string is rejected. There is no relation check to the build date.
- **FR-003**: `cheatsheet.tool_version` MUST be a string of 1–64 Unicode code
  points, single-line, without surrounding Unicode whitespace and without Unicode
  control, format (`Cf`), line- or paragraph-separator characters. Content is
  otherwise free (`10.3p1`, `2.6.0-FreeBSD`).
- **FR-004**: Invalid freshness fails with new code `VAL004`, positioned at the
  offending key where known, otherwise at the document start. The check is shared by
  source validation and the Hugo direct-IR boundary (`validate.Freshness`).
- **FR-005**: `VAL004` is appended to the public diagnostics list; no existing code
  changes meaning. Absent fields produce no diagnostic of any severity.
- **FR-006**: The Hugo renderer MUST expose the validated strings as `updated` and
  `tool_version` in the library and sheet data (omitted when absent) and MUST NOT
  call any time function, read a clock, or reformat the date.
- **FR-007**: Templates MUST render, in all three themes, shelf/index facts and
  sheet-header facts as static escaped text with the date inside
  `<time datetime="YYYY-MM-DD">`; the visible date text is the same ISO string.
  Labels are visible text (`Version`, `Updated`), so no extra ARIA name is needed
  and nothing relies on colour alone. No JavaScript, request or new font.
- **FR-008**: Freshness MUST appear in print output (no print-hiding rule), use
  existing palette tokens only, and wrap at 320px.
- **FR-009**: Shelf order, route, related scoring, tag pages and search content
  MUST NOT depend on the freshness fields.
- **FR-010**: Reference import MUST archive any upstream `cheatsheet.updated` and
  `cheatsheet.tool_version` inertly before validation with `IMP112`, merged with
  the existing `cheatsheet.cards` archive; it MUST NOT infer either field.
- **FR-011**: `smill build navi` MUST keep `.cheat` bytes, card identities and
  existing diagnostics unchanged; the version is not added to navi tags; the
  companion keeps the fields as generic metadata. Navi does not validate them.
- **FR-012**: The JSON Schema, `docs/format/gfm-plus.md`, `docs/diagnostics.md`,
  `docs/hugo.md`, `docs/imports.md`, `docs/format/navi-profile.md` and `docs/cli.md`
  describe the implemented contract.
- **FR-013**: The 15 sheets in `examples/sheets` gain real `updated` values and,
  where the sheet names a documented version, `tool_version`, derived and recorded
  as in [research.md](research.md); no `cheatsheet.tool` is added to them (it would
  add a navi tag).
- **FR-014**: Verification MUST keep every pre-existing parser/import/navi golden
  byte-identical, add one small freshness parser golden, and list every changed file
  of generated output (see [contract](contracts/freshness.md)).

### Key Entities

- **Freshness metadata**: optional author-stated update date and documented tool
  version; presentation and provenance facts only.
- **Freshness line**: the rendered text `Version X · Updated D`.

## Success Criteria

### Measurable Outcomes

- **SC-001**: A table of at least 40 date and 25 version cases yields exactly the
  documented accept/reject result and position, in source and direct-IR paths (FR-002..004).
- **SC-002**: Two real-Hugo builds with different `--clock` values and shuffled
  input order are byte-identical across all three themes (US1.5, FR-006).
- **SC-003**: Without the fields, the rendered HTML of all three themes is
  byte-identical to the base revision `05eba4d` outputs for existing fixtures (FR-007, US1.3).
- **SC-004**: All existing goldens and navi `.cheat` bytes are unchanged; the navi
  pair test shows identical `.cheat` bytes with/without the fields (FR-011, FR-014).
- **SC-005**: Hostile values appear only HTML-escaped; `datetime` attributes match
  `^[0-9]{4}-[0-9]{2}-[0-9]{2}$` in every emitted page (FR-007).
- **SC-006**: The demo library builds with the fields visible in the three themes
  and `smill check` passes on all 15 sheets (FR-013).

## Assumptions

- Authorized by the user on 2026-10-16 (feature and a draft PR), recorded in
  AGENTS.md. Merge, deployment and distribution remain separately gated.
- Browser E2E needs Node, which is not installed in the authoring environment; it
  is not claimed. Markup, determinism and CSS are covered by Go and Hugo tests.
- Out of scope: automatic dates, "stale" badges, relative dates ("3 days ago"),
  sorting or filtering by date, search over freshness, a future-date lint,
  localized date formats, sheet revision history, changelog and any manifest.
