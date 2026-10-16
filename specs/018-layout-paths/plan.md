# Plan: 018 layout paths

## Constitution check

- 1 Determinism: PASS. Same revision gives byte-identical trees (FR-8); the verifier
  only reads the tree and writes nothing; the shared fixed Hugo clock is kept.
- 2 Canonical source: PASS. No source format or schema change.
- 3 Loss-aware imports / 4 Inert input: PASS. No import or execution; the verifier
  only reads generated files.
- 5 Focused Go design: PASS with a justified exception. A POSIX shell script is the
  deploy gate in a Go-first repo; chosen because it is the smaller change (no new CLI
  surface, no new non-test package, runs on the built tree without a build step) and
  every rule is a text match a reviewer can read. Tests and the workflow parser are Go
  in the test-only package `internal/pagesworkflow` (`mvdan.cc/sh/v3` is already
  pinned in `go.mod`). Switching to Go stays possible if the rules outgrow `grep`.
- 6 Independent product: PASS. No sibling dependency.
- 7 Human publication control: BLOCKED for the workflow change. Two new public paths
  extend the 012 approval and need the user's confirmation (spec FR-12); the PR and
  topic push are authorized, deployment and merge are not. T005 and T007 wait.
- 8 Review evidence: PASS. Rounds in [analysis.md](analysis.md); commands and gaps
  reported.

1. Add `.github/scripts/verify-pages.sh` (POSIX shell with `grep -E`, `sed`, `find`,
   `sort`, `comm`; no network) and its table tests first, in `internal/pagesworkflow`
   (run through `os/exec`; skipped if `sh` is missing, a skip is not a pass).
2. Add the workflow test (fails on today's workflow), then edit `pages.yml`:
   replace the single demo export by three explicit exports, a three-way collision
   guard and mount, then the existing llms/agent-guide copy, then the verification script as the last step before upload.
3. Add the Hugo-gated integration test.
4. Update `specs/README.md` and AGENTS.md (T001, not blocked). Docs, llms.txt, agent guide
   and README (T007) are blocked on the user's confirmation, as is the workflow edit
   (T005, with its test T004).
5. Local verification: gofmt, vet, test, `make check`, `make test-hugo` with
   `SHEETMILL_HUGO=/opt/bundledock/toolbox/bin/hugo`. No workflow run, no deploy.

The build step keeps one `run` block so the existing `hugo()` helper and clock are
shared; verification is its own step, `Verify the Pages tree`, after it and before
upload. Exports are written out explicitly (no loop) so the test and reviewers can
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
