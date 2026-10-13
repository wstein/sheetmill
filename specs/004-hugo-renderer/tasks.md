# Tasks — 004

- [x] T001 Capture scope and dependencies.
- [x] T002 Record user authorization for brand finalization, implementation and local generation.
- [x] T003 Define the rendering, escaping, IO and Hugo tool contracts.
- [x] T004 Finalize approved SVG/raster brand suite with font provenance and visual QA.
- [x] T005 Write failing renderer/CLI safety, fidelity, determinism and output-boundary tests.
- [x] T006 Implement pure IR rendering and a self-contained original Hugo project.
- [x] T007 Add the functional Hugo export CLI and explicit new-directory handling.
- [x] T008 Build the local demo twice; verify actual bytes, desktop/mobile and search.
- [x] T009 Run required checks and independent Sol review; fix and re-review findings.
- [x] T010 Publish a reviewable feature PR and measured evidence (no deployment/merge).

Completed through [PR #3](https://github.com/wstein/sheetmill/pull/3).
See [verification evidence](../../docs/hugo-verification.md); no merge/deployment performed.

## Follow-up hardening (after PR #3)

Test gaps found in the post-merge review. Each needs a failing test first;
behavior fixes stay within the authorized 004 scope.

- [x] T011 Test that unknown frontmatter keys round-trip byte-stable into
  `data/ir/<sheet>.json` (FR-006).
- [x] T012 Inject a write failure partway through `writeProject`; assert `IO003`,
  exit status 2 and that no pre-existing path is modified (FR-003).
- [ ] T013 Keep a bare `[x](#)` link as `href="#"`; `url.String()` currently drops
  the empty fragment and renders `href=""`.
- [ ] T014 Run the Hugo integration test with `SHEETMILL_HUGO` on the REN103 and
  empty-label fixtures and record the evidence in docs/hugo-verification.md.
