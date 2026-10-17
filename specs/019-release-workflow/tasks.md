# Tasks: 019 release workflow

Write each test first and watch it fail. T004 fails until T005 exists, so they land in
one commit. No task runs a workflow, creates a tag or opens a PR.

- [ ] T001 [FR-9, authorization] `AGENTS.md`: add a feature-019 line quoting the owner (2026-10-17) with its exact scope and exclusions; amend the "publish artifacts from CI" rule to name the release workflow beside Pages; mark spec status.
- [ ] T002 [FR-2, FR-5, FR-4, FR-6] Failing script tests in `internal/releaseworkflow` (temporary git repos, `sh` via `os/exec`, skipped with a message if `sh`/GNU `tar` is missing; a skip is not a pass): verify (`version`, `tag`: lightweight, off-main, bad pattern, annotated ok; `go` mismatch), package (layout, modes, determinism, sidecar), smoke (good archive passes; wrong version, bad checksum, non-deterministic export fail), notes (first release, previous tag range, no marketing words).
- [ ] T003 Implement the four scripts (`# SPDX-License-Identifier: EUPL-1.2`, POSIX `sh`) until T002 passes.
- [ ] T004 **Lands with T005.** [FR-1..FR-8] Failing workflow tests: `violations()` baseline empty, the mutation table, job graph, permissions, pins, triggers, matrix equals check.yml's, ordering of verify steps, publish command flags.
- [ ] T005 Add `.github/workflows/release.yml` until T004 passes. Do not run it.
- [ ] T006 [FR-9] `check.yml` header comment; parity test (release matrix == check matrix) in `internal/releaseworkflow`.
- [ ] T007 [FR-10] Docs: `docs/releasing.md` (add it to nothing in `docs/manual-pages.txt` unless the review asks), README, `docs/getting-started.md`, `docs/cli.md`, `llms.txt`, `specs/README.md` row.
- [ ] T008 Local checks: gofmt, vet, test, `make check`, `make test-hugo` (`SHEETMILL_HUGO=/opt/bundledock/toolbox/bin/hugo`); a local dry run of the scripts against a locally built binary; record in analysis.md. Node is not installed, so `make test-browser*` is not run.
- [ ] T009 Independent review rounds recorded in `analysis.md`; push the topic branch; no PR, no workflow run.

Dependencies: T001 first; T002 -> T003; T004 + T005 together; T006 after T005; T007 after T005.
