# Tasks — 011

- [ ] T001 Obtain separate authorization for the 011 IR heading-contract change.
- [ ] T002 Decide additive field vs schema bump and the field shape; record it.
- [ ] T003 Write failing table-driven tests for the edge cases in spec.md.
- [ ] T004 Record Hugo fixture baseline bytes and current golden JSON.
- [ ] T005 Emit attribute-free heading nodes with exact spans in internal/markdown.
- [ ] T006 Remove suffix re-derivation from internal/render/hugo; render IR as given.
- [ ] T007 Confirm Reference import output and fixtures stay byte-identical.
- [ ] T008 Update model, JSON Schema, docs/format/gfm-plus.md, compatibility.md,
  hugo.md, architecture.md and docs/recommendations.md.
- [ ] T009 Add a fuzz target for heading titles and spans.
- [ ] T010 Regenerate goldens via `-run TestGolden -update`; inspect semantically.
- [ ] T011 Compare Hugo output byte-for-byte; justify every difference.
- [ ] T012 Run gofmt, go test ./..., go vet ./..., git diff --check.
- [ ] T013 Independent review; fix and re-review. No merge without human approval.
