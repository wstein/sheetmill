# Tasks: 018 layout paths

- [ ] T001 Record the authorization in `AGENTS.md` and the row in `specs/README.md`.
- [ ] T002 Failing table tests for `.github/scripts/verify-pages.sh` (valid tree and one failing tree per rule 1-7).
- [ ] T003 Implement the script until T002 passes.
- [ ] T004 Failing workflow test (`tests/pages`) for FR-1..FR-3, FR-6 and step order.
- [ ] T005 Edit `.github/workflows/pages.yml` until T004 passes. Do not run it.
- [ ] T006 `SHEETMILL_HUGO`-gated integration test: three real builds, script passes, repeat build byte-identical.
- [ ] T007 Docs: `docs/hugo.md`, `llms.txt`, `docs/agent-guide.md`, `README.md`.
- [ ] T008 Local checks: gofmt, vet, test, `make check`, `make test-hugo`; record in analysis.md.
- [ ] T009 Independent review rounds recorded in `analysis.md`; push topic branch. No PR, merge or deployment.
