# Plan: 018 layout paths

1. Add `.github/scripts/verify-pages.sh` (POSIX shell with `grep`, `find`, `sort`,
   `comm`; no network) and its table tests first.
2. Add the workflow test (fails on today's workflow), then edit `pages.yml`:
   replace the single demo export by three explicit exports, a three-way collision
   guard and mount, then run the script, then the existing llms/agent-guide copy.
3. Add the Hugo-gated integration test.
4. Update docs, llms.txt, agent guide, README, `specs/README.md`, AGENTS.md.
5. Local verification: gofmt, vet, test, `make check`, `make test-hugo` with
   `SHEETMILL_HUGO=/opt/bundledock/toolbox/bin/hugo`. No workflow run, no deploy.

The workflow keeps one `run` block so the existing `hugo()` helper and clock are
shared. Exports are written out explicitly (no loop) so the test and reviewers can
read each command; the duplication is three lines.

Exact commands (after `$RUNNER_TEMP/manual` export and the manual Hugo build):

```
go run ./cmd/smill build hugo examples/sheets --out "$RUNNER_TEMP/demo" ... --base-path /sheetmill/demo/ --site-url https://wstein.github.io/sheetmill/demo/ --header-link 'Compact layout=/sheetmill/demo-compact/' --header-link 'Cards layout=/sheetmill/demo-cards/'
go run ./cmd/smill build hugo examples/sheets --out "$RUNNER_TEMP/demo-compact" ... --theme classic-compact --base-path /sheetmill/demo-compact/ ...
go run ./cmd/smill build hugo examples/sheets --out "$RUNNER_TEMP/demo-cards" ... --theme reference-cards --base-path /sheetmill/demo-cards/ ...
```

Risks: label typos or a missing trailing slash break links silently; the script's
rules 2, 3 and 6 and the workflow test cover them. A manual page named like a
variant is caught by FR-6.
