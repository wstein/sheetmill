# Tasks — 006

Documentation-only refinement; unchecked implementation tasks remain gated.

- [x] T001 Capture renderer scope and dependencies.
- [x] T001a Draft safety rules and diagnostics (superseded by the refined profile).
- [x] T001b Propose explicit cards, card-keyed metadata, free-input parameters and
  one `bash-free-input-v1` profile; document format and immutable upstream evidence.
- [ ] T002 Obtain scope-specific implementation authorization; this docs PR grants none.
- [ ] T003 Accept the refined contract through independent review/human merge;
  confirm the acceptance matrix. This is design acceptance, not a passed test claim.
- [ ] T004 After authorization, pin a pure Bash grammar parser, define reviewed
  explicit-ID IR provenance and sidecar schema, and write failing acceptance tests,
  golden bytes, pure pinned-navi parser/interpolation fixtures, determinism checks,
  fuzz targets and transitive no-execution audit. No navi/shell invocation.
- [ ] T005 Implement selection/metadata validation, renderer and proposed CLI in
  a focused branch; update JSON Schema, NAV diagnostics, docs/cli.md and user guide.
  Keep default/enum/provider/multiline/other-shell capabilities out of scope.
- [ ] T006 Obtain generation/build approval before producing output artifacts.
- [ ] T007 Run the complete acceptance matrix, record exact compatibility evidence
  and obtain independent review before human merge. No source commands execute.
- [ ] T008 Obtain separate publication approval if requested.
