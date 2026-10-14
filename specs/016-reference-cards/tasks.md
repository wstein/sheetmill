# Tasks: reference-cards

- [x] T001 Record the authorization in `AGENTS.md`.
- [x] T002 Failing tests in `internal/render/hugo/theme_cards_test.go`: theme accepted, `cards.css` fingerprinted only for reference-cards, `cards` param, editorial/classic-compact output has no cards assets, deterministic output.
- [x] T003 Implement the theme option, params and asset handling in `internal/render/hugo/hugo.go`; list the value in `Themes` and the CLI flag text.
- [x] T004 Templates: body class and stylesheet link in `baseof.html`; Browse disclosure instead of the sidebar; centered home heading.
- [x] T005 `cards.css`: tiles, badge, rows, `$` prompt, headings, narrow/print rules from FR-3..FR-9, using palette tokens only.
- [x] T006 Real-Hugo integration test: body class, Browse panel with 014 order and `aria-current`, no sidebar, both a home and a sheet page.
- [x] T007 Docs: `docs/hugo.md`, `docs/cli.md`, `specs/README.md`.
- [x] T008 Browser check (Chrome) at desktop, 390px and 320px, light/dark; print was not previewed; record limits honestly.
- [ ] T009 Final readiness: Go/Hugo tests, vet, diff checks and the 51-journey reader/corpus suite pass after review fixes. Exact-head Sol re-review and current-head CI must be clear before leaving draft. Do not merge or deploy.

T008 is now covered by the pinned Playwright harness as well as the original Chrome
check. Six new root/subpath journeys exercise desktop, mobile and no-JavaScript;
compact/comfortable state is asserted where JavaScript provides the toggle. Tests
cover centering, native navigation, palette contrast, exact copying without the
CSS prompt, 320–1280px widths, badge flow, readable table scrolling and print-media
wrapping. No-JavaScript mode tests its default compact layout and system dark mode.

The initial author check did not use the pinned harness. The subsequent review
ran all 51 reader/corpus journeys successfully; see [review evidence](review-evidence.md).
Print **media emulation** was checked, not physical printing, PDF pagination,
Firefox or WebKit. T009 stays open until PR-level review/CI readiness is recorded.

Historical author review cleared `fdee8ff` after an earlier fix round. Concurrent
follow-ups `5815002` and `74f0291` then strengthened print rules, corner/tint
selectors and contrast tests. Those changes are preserved in this review branch.
T009 is pending for the combined final head rather than reusing an older CI result.
