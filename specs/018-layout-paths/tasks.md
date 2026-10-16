# Tasks: 018 layout paths

- [ ] T001 Record the authorization in `AGENTS.md` (amend the 012 bullet, FR-12), specs/012 (status, FR-009, Resolved), `docs/hugo.md` Pages text and the `specs/README.md` row.
- [ ] T002 Failing table tests for `.github/scripts/verify-pages.sh` (valid tree and one failing tree per rule 1-7).
- [ ] T003 Implement the script until T002 passes.
- [ ] T004 Failing workflow test (`internal/render/hugo`-independent package `internal/pagesworkflow`, parsing `run:` with `mvdan.cc/sh/v3`) for FR-1..FR-3, FR-6 and step order.
- [ ] T005 Edit `.github/workflows/pages.yml` until T004 passes. Do not run it.
- [ ] T006 `SHEETMILL_HUGO`-gated integration test in `internal/pagesworkflow`; add that package to the `test-hugo` Makefile target so the gate runs it: three real builds, script passes, repeat build byte-identical.
- [ ] T007 Docs: `docs/hugo.md`, `llms.txt`, `docs/agent-guide.md`, `README.md`.
- [ ] T008 Local checks: gofmt, vet, test, `make check`, `make test-hugo`; record in analysis.md.
- [ ] T009 Independent review rounds recorded in `analysis.md`; push the topic branch (the coordinator opens the draft PR). No merge or deployment.
