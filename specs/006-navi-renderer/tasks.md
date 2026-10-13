# Tasks — 006

- [x] T001 Capture renderer scope and dependencies.
- [x] T001a Draft safety rules and diagnostics (superseded by the refined contract).
- [x] T001b Propose explicit cards, card-keyed metadata, free-input parameters and
  pinned upstream evidence (navi v2.24.0).
- [x] T002 Implementation authorization (user, 2026-10-14; see AGENTS.md).
- [x] T003 Resolve decisions: file format, explicit-ID selection, shell profiles,
  warnings for unexportable cards, no navi execution.
- [x] T004 Failing-first tests: negative table, golden bytes, grammar model with
  entry deduplication, fuzz target, no-execution lint; additive IR `id_explicit`.
- [x] T005 Implement `internal/render/navi` and `smill build navi`; document in
  docs/format/navi-profile.md, docs/cli.md and docs/architecture.md.
- [ ] T006 Obtain generation/build approval before producing published artifacts.
- [ ] T007 Independent review of the final revision before human merge. Author
  checks are not review.
- [ ] T008 Obtain separate publication approval if requested.
- [ ] T009 Metadata sidecar (`<sheet>.navi.json`) preserving non-navi metadata.
- [ ] T010 Pin a pure Go shell parser for bash/zsh/fish syntax checks.
- [ ] T011 Later: choice lists, shared contexts, nushell raw strings.
