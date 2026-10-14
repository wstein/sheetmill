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
- [x] T007 Exporter candidate independently reviewed at `caf1fc6` (see spec delivery
  evidence); integration followed through #27/#28. This does not approve new commits.
- [ ] T008 Obtain separate publication approval if requested.
- [x] T009 Metadata sidecar (`<sheet>.navi.json`) preserving complete metadata.
  Implemented in [#30](https://github.com/wstein/sheetmill/pull/30): schema-1
  complete semantic metadata, deterministic bytes, actual exported-card identities,
  strict/export boundaries and batch failure tests. Final independent exact-revision
  review is recorded on that PR; no navi execution or publication approval implied.
- [ ] T010 Stronger shell syntax checks. Bash/Zsh quote diagnostics now use only
  pinned `mvdan.cc/sh/v3/syntax` v3.14.1; full syntax validation and a suitable
  Fish parser remain pending.
- [ ] T011 Later: choice lists, shared contexts, nushell raw strings.
