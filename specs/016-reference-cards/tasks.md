# Tasks: reference-cards

- [x] T001 Record the authorization in `AGENTS.md`.
- [x] T002 Failing tests in `internal/render/hugo/theme_cards_test.go`: theme accepted, `cards.css` fingerprinted only for reference-cards, `cards` param, editorial/classic-compact output has no cards assets, deterministic output.
- [x] T003 Implement the theme option, params and asset handling in `internal/render/hugo/hugo.go`; list the value in `Themes` and the CLI flag text.
- [x] T004 Templates: body class and stylesheet link in `baseof.html`; Browse disclosure instead of the sidebar; centered home heading.
- [x] T005 `cards.css`: tiles, badge, rows, `$` prompt, headings, narrow/print rules from FR-3..FR-9, using palette tokens only.
- [x] T006 Real-Hugo integration test: body class, Browse panel with 014 order and `aria-current`, no sidebar, both a home and a sheet page.
- [x] T007 Docs: `docs/hugo.md`, `docs/cli.md`, `specs/README.md`.
- [x] T008 Browser check (Chrome) at desktop, 390px and 320px, light/dark; print was not previewed; record limits honestly.
- [ ] T009 (in progress: review round 1 NOT CLEAR at fff1223, fixes in the next commit) gofmt, vet, full tests with Hugo, race, diff check; independent review of the exact head; fix findings; draft PR, then mark ready only after CI and review. Do not merge or deploy.

T008 evidence: Chrome check of the built `examples/sheets` demo (desktop screenshot, iframe measurements at 320px and 390px with the Browse panel open: no horizontal overflow, no badge overflow; dark badge uses `--link` on `--bg`). Real print and Firefox/WebKit were not tested; the pinned Playwright suite was not run locally (Node 26 installed, 24.21.0 required).
