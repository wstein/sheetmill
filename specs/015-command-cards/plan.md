# Plan — 015 command cards

Status: docs-only design, not implementation authorization. Baseline main `f12a4b7`.

## Architecture decisions for a later handoff

- Keep generic schema-1 metadata and raw heading classes; derive command semantics
  without mutating source IR. Add the parser's `.command` whitelist entry later.
- Core owns shared `cheatsheet.cards` shape/type rules and document-level explicit
  ID/reference rules. Metadata-only validation cannot check card references; run
  that stage after Markdown structure exists. Direct renderer consumers share it.
- Preserve the legacy validation boundary: `cheatsheet.navi` remains renderer-owned
  and legacy missing-ID selection retains warning/skip behavior. Do not impose
  new core errors on legacy-only documents as an incidental migration change.
- Source diagnostics use nested key/card spans; direct IR falls back to 1:1.
  Reserve CMD001/CMD002 only in proposal docs now; add code/schema/tests later.
- Navi consumes derived command semantics and resolves compatibility aliases once.
  Compare complete normalized per-card declarations, never use map iteration order
  or silently overwrite one namespace. Invalid legacy settings still fail NAV006.
- Shared metadata does not decide navi identity. Keep the existing `context`,
  `tagList`, `oneLine`, `itemHash`, format version and collision checks unchanged.
- Companions serialize original metadata, not the normalized alias view. Old/new
  sidecar bytes can differ; exported-card identity arrays must agree.
- Reference imports must archive upstream `cheatsheet.cards` and remove/archive
  upstream `.command` before canonical validation/selection. Preserve unrelated
  cheatsheet fields and existing legacy import behavior; do not infer commands.

## Evidence inspected

`internal/render/navi/navi.go`: `renderSheet` currently reads `cheatsheet.navi`,
selects `.navi` and constructs prehash as `tagLine + oneLine(card.Title)`.
`context` derives `sheetmill:<encoded path>#<encoded id>!`.
`itemHash` uses the pinned v2.24.0 model; existing fixed vectors are in navi_test.go.
Do not replace it with a generic FNV implementation: preserve its exact seed,
UTF-8 bytes and trailing 0xff behavior and the committed vectors.
`internal/validate/validate.go` currently treats cards as generic metadata;
`internal/markdown/markdown.go` rejects `.command` today. `frontmatter.Parse`
returns all metadata diagnostics as fatal. Import validation precedes Markdown
structure, so both shared-field and marker archival timing need explicit tests.

## Future verification

Paired legacy/new/dual/mixed fixtures, per-entry expected hash baselines, complete
`.cheat` byte comparison, real collision NAV008 cases, core/direct-IR negative
cases, imports and companion preservation. Keep old goldens unchanged. Use Go and
the existing CLI test harness, never execute navi or source commands.

Constitution check: deterministic output, canonical source, inert input and
publication boundaries retained. No runtime dependencies or rendering UI added.
Implementation authorization, failing-first tests and exact-revision independent
review are prerequisites to a future implementation PR.
