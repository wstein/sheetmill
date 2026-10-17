# Plan: 019 release workflow

## Constitution check

- 1 Determinism: PASS. Archives and binaries are checked byte-identical (FR-3, FR-5); the
  smoke export runs twice and compares.
- 2 Canonical source: PASS. No source format change.
- 3 Loss-aware imports / 4 Inert input: PASS. The smoke test only checks and exports the
  bundled example sheets; nothing in a sheet executes; no Hugo run.
- 5 Focused Go design: PASS with the 018 exception. POSIX shell scripts and one YAML
  workflow carry the release steps; the tests are Go in a test-only package using the
  already pinned `go.yaml.in/yaml/v3` and `mvdan.cc/sh/v3`. No new dependency, no new CLI
  surface, no non-test Go code.
- 6 Independent product: PASS.
- 7 Human publication control: PASS conditionally. The owner's 2026-10-17 decision is the
  recorded approval for tag-triggered publication; the spec grants nothing itself. The
  constitution's own wording names only Pages as an approved named workflow; amending it is
  the owner's call (spec open decision 1). Merge, tags, workflow runs and environment
  rules stay human.
- 8 Review evidence: PASS. Rounds in [analysis.md](analysis.md).

## Files

| File | Purpose |
| --- | --- |
| `.github/workflows/release.yml` | jobs `verify`, `build`, `smoke`, `publish` |
| `.github/scripts/release-verify.sh` | subcommands `version`, `tag`, `go` (git checks; no network) |
| `.github/scripts/release-package.sh` | deterministic archive + `.sha256` sidecar |
| `.github/scripts/release-smoke.sh` | extract and exercise a shipped archive |
| `.github/scripts/release-notes.sh` | release notes text on stdout |
| `.github/workflows/check.yml` | header comment only |
| `internal/releaseworkflow/` | tests: workflow rules + mutations, scripts, check/release parity |
| `docs/releasing.md`, README, `docs/getting-started.md`, `docs/cli.md`, `llms.txt`, `specs/README.md`, `AGENTS.md` | docs and authorization |

Network- and `gh`-dependent steps (tag fetch, check-run query, release existence, ls-remote,
`gh release create`) stay inline in the workflow, small and covered by structure tests; all
logic that can run offline lives in scripts that are tested against temporary git
repositories.

## Job graph

```
verify --> build (3 targets) --> smoke (3 native runners) --> publish (tag push only)
```

`verify` outputs `version`, `dry_run`, `commit`. `build` uploads `release-<os>-<arch>`
(archive and sidecar, `retention-days: 1`, `if-no-files-found: error`). `smoke` and
`publish` download them. `publish` runs only if
`needs.verify.outputs.dry_run == 'false' && github.event_name == 'push' && github.ref_type == 'tag'`.

## Script contracts

- `release-verify.sh version <v>`: exit 0 iff `v` matches the strict release regex.
  `release-verify.sh tag <tag> <main-ref>`: strict regex; `git cat-file -t refs/tags/<tag>`
  is `tag`; `git rev-parse <tag>^{commit}` is an ancestor of `<main-ref>`. `release-verify.sh
  go`: `.go-version` equals the `go` line of `go.mod`. Each failure prints one line to stderr.
- `release-package.sh <version> <os> <arch> <binary> <outdir>`: reads `LICENSE`, `README.md`
  from the working directory, `SOURCE_DATE_EPOCH` from the environment (required), writes
  the archive and `<archive>.sha256` into `<outdir>`.
- `release-smoke.sh <version> <archive> <repo-root>`: verifies the sidecar checksum if
  present, extracts into a private temp dir, runs the FR-4 commands.
- `release-notes.sh <tag> <commit> <go-version>`: stdout only.

## Order (TDD)

1. T001 `AGENTS.md` authorization line, spec status.
2. T002/T003 script tests then scripts.
3. T004/T005 workflow tests (fail: no file) then `release.yml`; the failing test and the
   workflow land in one commit so `go test ./...` stays green on every commit, as in 018.
4. T006 `check.yml` comment and parity test.
5. T007 docs. T008 local checks. T009 review, push (no PR).

## Risks

A typo in a path filter or a job `needs` silently disables a gate; the mutation tests pin
them. The tag path cannot run before the first tag; the dry run covers build/smoke only and
the first real run needs the owner watching (documented). `ubuntu-24.04-arm` or `macos-15`
availability changes would block releases; fail closed.
