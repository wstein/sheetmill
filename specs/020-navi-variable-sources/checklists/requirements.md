# Specification quality - 020 navi variable sources

Created 2026-10-20. Design quality and cross-artifact consistency only; no
implementation or runtime test has been run.

- [ ] User stories cover choices, sources with options, and unchanged existing sheets.
- [ ] Every FR is testable and has a task and a named test.
- [ ] SC-001..SC-005 map to tasks T007/T008, T001/T012, T002/T010, T006/T013, T014.
- [ ] Edge cases (empty, duplicate, quotes/newlines/unicode choices; unbalanced source; both fields; unused parameter; per-shell quoting; fish and nu) appear in the contract's negative table.
- [ ] IR change is stated as none; schema versions, entry hashes and NAV008 are stated unchanged.
- [ ] The trust model is recorded as a decision (no flag, no gating), and the no-execution statement appears only as a scope/boundary note.
- [ ] Dependent variables are out of scope and listed as a later feature.
- [ ] navi syntax claims are sourced in research.md; unverified points are marked to-be-measured, none invented.
- [ ] Diagnostics continue the sequence (`NAV108`, `NAV109`) and extend `CMD001`/`NAV006`.
- [ ] No tool or navi version is hard-coded in tests; documents point to existing evidence.
- [ ] Cross-check: golden examples in the contract match the quoting table and the line invariants.
- [ ] Cross-check: 006 text that says `$` lines are never emitted is listed for update in T016.
- [ ] Cross-check: PR sizing, TDD order and the AGENTS.md grant are consistent.

## Analysis notes

- Ambiguity resolved: option strings are single-quoted and a value with `'` is
  rejected for simplicity; navi's splitter would parse `'\''`, but SheetMill does not rely on it.
- Decision for review: duplicate choices are an error (not silently merged).
- Decision for review: shared limits (64 choices, 128 code points) are proportionate
  caps and can be raised; longer lists belong in a `source`.
