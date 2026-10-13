# 005 — Typst renderer

Created: 2026-10-13 · Status: draft, specification only · Depends on: 002-gfm-parser
and accepted 001–004 (IR schema 1, diagnostics, URL safety from `REN002`)

Statements marked "to verify" rest on general knowledge of Typst and must be
checked against the pinned version before the contract is accepted.

## User scenarios and acceptance

1. P1: An author runs `smill build typst <file-or-directory>... --out <new-dir>` and gets
   Typst sources for multi-page PDF references from canonical IR. Compiling the PDF
   with the pinned Typst is a separate, separately authorized step (T006).
2. P1: Given identical IR, configuration, fonts and pinned Typst, the `.typ` bytes and
   the PDF bytes are identical across runs, input orders and hosts.
3. P1: Hostile text never becomes Typst markup or code. Unsupported or unsafe content
   yields source-positioned diagnostics; nothing disappears silently.
4. P2: A reader gets a cover/index, per-sheet sections, a source/license page and
   cards that stay whole on a page.

Acceptance scenarios (Given/When/Then):

- A1: Given text `*a* #let x = 1 \ $ @ < > ` and `#import "x"`, when rendered, then
  each character is emitted escaped and the output compiles to the same literal text
  (fixture compared by bytes; compile check is post-T006).
- A2: Given a raw HTML node, when rendered, then it is shown as escaped literal text
  in a code-styled block and `TYP101` is reported; no HTML is interpreted.
- A3: Given a code block containing `` ``` ``, `#`, `\` and `*/`, when rendered, then
  the content is emitted verbatim with a fence longer than any backtick run in it.
- A4: Given a link `javascript:alert(1)` or `//host`, when rendered, then the build
  fails with `REN002` and no output directory is created.
- A5: Given `![alt](https://example.org/x.png)`, when rendered, then no `image()` call
  is emitted, the image is retained as alt text plus an unfetched link, and `TYP102`
  is reported; no network access occurs.
- A6: Given an IR node of unknown kind, when rendered, then the build fails with
  `REN001` (error), never skipping the node.
- A7: Given a card taller than one page, when rendered, then it breaks across pages
  at block boundaries and `TYP103` (warning) is reported once per card.
- A8: Given the same inputs in reversed order and a different working directory,
  when rendered twice, then all output bytes are identical.
- A9: Given a configured Typst whose `typst --version` differs from the pinned
  version, when the compile step runs, then it refuses with `TYP002`.
- A10: Given a font profile whose file hash mismatches or whose file is missing,
  when rendered or compiled, then it fails with `TYP003`; no host font is used.
- A11: Given an existing or symlinked output path, when rendered, then the build is
  refused with `IO003` and nothing is written or removed.
- A12: Given unknown frontmatter keys, when rendered, then they are carried in
  `metadata.json` next to the sources; none is dropped.

## Requirements

- FR-001: Consume schema-1 IR directly; never reparse Markdown or use Typst's
  own Markdown/evaluation features on source text.
- FR-002 (escaping contract): Every IR text node (headings, titles, table cells,
  link text, alt text, metadata strings, attribution) is emitted through a single
  escape function that neutralises all Typst markup and code characters
  (to verify: `\ # $ * _ ` ~ < > @ [ ] { } ( ) = - + / " ' :` and line-start markers);
  where escaping is not sufficient, text is emitted as a string literal with `\\`,
  `\"`, `\n`, `\u{..}` escapes. Generated sources contain no `#import`, `#include`,
  `#read`, `#eval`, `#plugin` or `#image` call derived from IR; a test scans every
  fixture output and fails on these. Code is emitted verbatim in a raw block whose
  fence is longer than any backtick run in the content. Raw HTML is inert escaped
  text with `TYP101`. Language tags of code blocks are allowlisted identifiers or dropped.
- FR-003: `smill build typst ... --out <new-dir>` requires a new directory; refuse
  existing/symlink destinations, validate all inputs before writing, no overwrite
  mode, no implicit network/tool downloads (as 004 FR-003/FR-008).
- FR-004 (page model; values are proposals to confirm in T003):
  - Page size: `a4` default, `letter` option; portrait; single column body, optional
    two-column card grid in the reference profile.
  - Margins: 18 mm, fixed per profile; no per-sheet override.
  - Cards (`###`) use non-breaking blocks (`breakable: false`) with the heading kept
    with the first content block (to verify the exact Typst primitives). A section
    heading is never the last item on a page.
  - A card that cannot fit one empty page is allowed to break at block boundaries
    (never inside a code line or table row) and `TYP103` is reported.
  - Web layout hints (`col-span`, `row-span`) never influence pagination.
  - Page numbers, sheet title in the header and sheet-level outline are mandatory.
- FR-005 (determinism): Fixed document creation date and metadata derived from
  inputs only (no clocks, no host paths, no usernames); stable ordering by sheet
  path; tool pinned in a single source file `.typst-version`, checked in CI;
  the compile step refuses any other `typst --version` output with `TYP002`
  (build suffixes handled as for Hugo). Whether Typst accepts a fixed creation
  timestamp and yields identical PDF bytes across hosts is to verify against the
  pinned version; if not, compare `.typ` bytes and extracted PDF text/structure.
- FR-006 (fonts): Fonts are embedded only from local, licensed, hash-pinned files
  with licenses under `assets/licenses`. Today `assets/fonts` holds only a WOFF2
  (IBM Plex Sans Regular) for web; Typst reads TTF/OTF/TTC (to verify, WOFF2 not
  expected to work). Therefore a PDF font profile (sans, mono, bold/italic as
  needed; see `assets/BRAND.md`) is an open decision that needs new licensed
  assets with provenance before T004. Compile uses explicit font paths and ignores
  system fonts (to verify: pinned Typst flag for this, such as `--ignore-system-fonts`).
  Missing glyphs are reported (`TYP003` for profile errors; fallback is never silent).
- FR-007 (accessibility): Target is an open decision. Candidate: tagged PDF with
  document language, title, heading structure, alt text and a logical reading order,
  and PDF/UA conformance if the pinned Typst supports it (to verify; support
  and flags vary by version). The minimum for T007 review: document title and
  language set, headings are real headings, contrast of the coral/periwinkle
  palette meets WCAG 2.2 AA for text, and a checker run (tool to select, to verify)
  is recorded without claiming conformance not reached.
- FR-008 (links): Same URL policy as `REN002`: relative, http, https, mailto only;
  rejected forms as listed in 004 FR-004. Unresolved internal links keep their text
  with warning `REN103`. Images are never fetched or embedded from sources.
- FR-009 (attribution): A final source/license page lists, per sheet, source path,
  title, license and provenance (from IR metadata and `docs/provenance.md` rules);
  each sheet's footer or first page names its source. Missing license metadata is
  rendered as "not stated" and warned, not omitted.
- FR-010: Unknown IR node kinds or properties fail or are diagnosed; unknown
  metadata is preserved (`metadata.json`) or diagnosed.
- FR-011: No snippet, shortcode, package (`@preview`) or linked resource is
  evaluated or fetched; generated sources use no Typst packages.
- FR-012: Diagnostics proposed range, documented in `docs/cli.md` and a new
  `docs/typst.md` before authorization:
  - `TYP001` (error): invalid or unsupported renderer configuration/profile.
  - `TYP002` (error): Typst version differs from the pinned version.
  - `TYP003` (error): font profile file missing, hash mismatch or unlicensed.
  - `TYP101` (warning): raw HTML shown inertly.
  - `TYP102` (warning): image retained as unfetched link.
  - `TYP103` (warning): card exceeds one page and breaks.
  - Shared `REN001`, `REN002`, `REN003`, `REN103`, `IO003` are reused, not renumbered.
- FR-013: Obtain separate implementation, build/PDF generation and publication
  authorization; this specification grants none.

## Success criteria

After authorization: tests for A1–A12 pass; golden `.typ` fixtures are byte-for-byte
and independently authored; table-driven escaping tests plus a fuzz target over
arbitrary text/code/URL inputs assert no forbidden constructs in output; repeated
generation yields identical hashes. A human inspects the generated PDF at least
at A4 and Letter, with long cards, tables and code; an independent reviewer is recorded
in a verification document; findings are fixed and re-reviewed. No build or
publication happens under the current milestone.

## Open decisions

Confirm page defaults (FR-004); the PDF font profile and its licensed TTF/OTF
assets (FR-006); accessibility target and checker (FR-007); whether fixed-date PDF
bytes are reproducible with the pinned Typst (FR-005); Typst version to pin.

## Authorization prerequisite

Measurable acceptance criteria (A1–A12, FR-002 to FR-012) now exist, but the open
decisions above must be closed and accepted, and fonts with provenance added,
before this feature can be authorized.
