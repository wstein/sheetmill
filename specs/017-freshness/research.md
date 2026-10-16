# Research — 017

## Decisions

1. **Field names.** `cheatsheet.updated` follows the user's suggestion. The tool
   version is `cheatsheet.tool_version`, not `version`: a bare `version` reads as
   the sheet's own version, and `cheatsheet.schema` already means something else.
   `cheatsheet.tool` already exists (name only); it is reused, not extended.
2. **Namespace stays open.** Both live under the existing generic `cheatsheet`
   mapping with per-field validation, like `tool`; no closed namespace, so no new
   source-compatibility break except for a previously unvalidated `updated` or
   `tool_version` value of another shape under `cheatsheet` (documented, as 014 did).
3. **Date shape.** Date-only ISO 8601 `YYYY-MM-DD`: unambiguous, sortable, maps 1:1
   to `<time datetime>`, and survives YAML timestamp-to-string handling already in
   `internal/frontmatter` (`!!timestamp` is kept as its source spelling). Parsing
   uses `time.Parse("2006-01-02")` after an ASCII-digit pattern check; the Go
   parser rejects impossible days and handles leap years. Year floor 1970 catches
   typos such as `0026`; year ceiling 9999 is the format limit.
4. **No future check.** Rejecting "after the build" needs a clock, which breaks
   constitution 1 (identical bytes) and the Pages build's fixed Hugo `--clock`
   (`2000-01-01`; any future relative to it would also reject every real date).
   A lint comparing against an explicit `--today` is a possible later feature.
5. **Display is the ISO string.** No month names or locale: no locale tables,
   no ambiguity, same text for `datetime`. A reader on any locale sees `2026-10-14`.
6. **New code `VAL004`.** `VAL003` was already widened for `cheatsheet.*` and
   navigation; a separate code lets CI separate "freshness mistakes" from
   structural ones. One code for both fields; the message names the field. Codes
   are append-only; `VAL003` text is unchanged.
7. **Import: archive, do not interpret.** Following 014/015, upstream values of
   either name are removed from active metadata before validation and archived
   under `source.original_metadata.cheatsheet.*`, finding `IMP112`. Reason: the
   semantics of upstream `updated` are unknown, and a failing import is worse than
   an inert archive. Merging into the existing `original["cheatsheet"]` map is
   required so `cards` archival (IMP111) keeps working.
8. **Navi.** `renderSheet` reads only `cheatsheet.tool` and `cheatsheet.navi`;
   the companion serializes all metadata. So bytes are unchanged by construction;
   a regression test pins it. The version is deliberately not a navi tag.
9. **Rendering shape.** One small helper partial `freshness.html` renders
   `<span class="fresh">`/`<p class="…-fresh">` with `Version …` and `Updated
   <time>` and a CSS-drawn `·` separator (decorative, not selectable text). Same
   partial for editorial card, classic index row and all sheet headers keeps wording
   consistent. Existing sheets without fields render exactly as before.
10. **Print and a11y.** The facts are text in normal flow, so they print by default
    and are read in order by assistive technology; the shelf card is one link whose
    accessible name already includes its content. `<time>` adds machine meaning
    without changing the spoken text.

## Demo data derivation

No sibling repository or network source was consulted (policy: no fetching).

- `updated`: the sheet carries no pinned revision date, so it is the committer date
  of the sheet file's last change at base `05eba4d` from
  `git log -1 --format=%cs -- <file>`: 2026-10-16 for bdock, reference-writing,
  sheetmill, smill, whr; 2026-10-14 for the other ten. It states when the sheet was
  last updated, not when the tool was released.
- `tool_version`: taken from each sheet's own `source.note` (the sheet records the
  documentation it was authored from): curl 8.7.1, git 2.56.0, jq 1.8.2, make 3.81
  (GNU Make), grep 2.6.0-FreeBSD, ssh 10.3p1 (OpenSSH), tar 3.5.3 (bsdtar). Omitted:
  find (macOS BSD, no version), regex (POSIX documentation), markdown (SheetMill
  source format, not a tool), bdock, sheetmill, smill, whr (no pinned release).
  These are verified only against the sheets' own text; no upstream was fetched.
