# Tasks: 018 layout paths

T005 and T007 are **blocked until the user confirms the new public paths** (spec
Authorization, FR-12). Everything else is test or script work on the topic branch.
Write each test first and watch it fail.

- [ ] T001 [FR-12] Record the proposed extension as pending the user's confirmation, not as granted: `AGENTS.md` (012 bullet), specs/012 (status, FR-009, Resolved) and the `specs/README.md` row. The confirmation line with its date is added only after the user confirms.
- [ ] T002 [FR-7] Failing table tests for `.github/scripts/verify-pages.sh` in `internal/pagesworkflow`, run with `sh` through `os/exec` (skip with a message if `sh` is missing; a skip is not a pass): one valid tree and one failing tree per rule 1-8, including the loopback prose link `http://127.0.0.1:1313/` (passes), a `//` src, an `http:` src/link, `example.invalid` in output, a manual slug `demonstration` (rule 1/3 reserved prefix), a wrong layout marker, a page-set difference and a self header link.
- [ ] T003 [FR-7] Implement the script (`sh`, `grep -E`, `sed`, `find`, `sort`, `comm`) until T002 passes.
- [ ] T004 [FR-1, FR-2, FR-3, FR-6, FR-7, FR-9] Failing workflow test in `internal/pagesworkflow`, parsing the YAML and each `run:` with `mvdan.cc/sh/v3`: the three exports, base-path/site-url agreement, theme flags, titles, header-link rules, shared clock and flags, collision guards, mount order, the separate `Verify the Pages tree` step after the llms/agent-guide copy and before upload (step order asserted), and unchanged triggers, permissions and `needs: build`.
- [ ] T005 [FR-1..FR-3, FR-6, FR-7] **Blocked on confirmation.** Edit `.github/workflows/pages.yml` until T004 passes. Do not run it.
- [ ] T006 [FR-4, FR-5, FR-7, FR-8] `SHEETMILL_HUGO`-gated integration test in `internal/pagesworkflow`, added to the `test-hugo` Makefile target: build the real manual from `docs/manual-pages.txt` and the three real demos through `internal/cli.Run`, mount as the workflow does, run `verify-pages.sh` (must pass, including the `http://127.0.0.1:1313/` prose link of `docs/hugo.md`), and check a repeat build is byte-identical.
- [ ] T006a [FR-7] Always-run template-text test in `internal/render/hugo`: every `href`, `src` and `action` attribute in `theme/layouts` is double-quoted (fails on single-quoted or unquoted).
- [ ] T007 [FR-13] **Blocked on confirmation.** Docs: `docs/hugo.md` (Pages and Search engines), `llms.txt`, `docs/agent-guide.md`, `README.md`; list the layout URLs and the accepted limitations.
- [ ] T008 Local checks: gofmt, vet, test, `make check`, `make test-hugo`; record in analysis.md.
- [ ] T009 Independent review rounds recorded in `analysis.md`; push the topic branch (the coordinator opens the draft PR). No merge or deployment.

Dependencies: T002 -> T003; T004 -> T005; T006 needs T003 and T005; T007 needs T005.
