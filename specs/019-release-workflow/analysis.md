# Analysis: 019 release workflow

## Coverage (FR/SC to tasks)

| Requirement | Tasks |
| --- | --- |
| FR-1 triggers | T004, T005 |
| FR-2 verify | T002, T003 (git checks), T004, T005 (API checks, order) |
| FR-3 build | T004, T005 |
| FR-4 smoke | T002, T003, T004, T005 |
| FR-5 archives | T002, T003, T004 |
| FR-6 publish | T002 (notes), T004, T005 |
| FR-7 security | T004 (mutations) |
| FR-8 tests | T002, T004, T006 |
| FR-9 pipeline | T001, T006 |
| FR-10 docs | T007 |
| SC-001..006 | T004 (SC-001, 005, 006), T002 (SC-002, 003, 004) |

Unverified until the first real tag: annotated-tag fetch, check-run API query, release
existence check, `ls-remote`, environment gate, `gh release create`, runner labels.
Not run by the author: any workflow, any tag operation.

## Review rounds

Order note: during review round 1 the author drafted the script tests, scripts and workflow
in the working tree (uncommitted, outside the reviewed commit). The script and workflow tests
were written before their scripts/workflow and seen failing (missing files); the commits
combine each test with its implementation so every commit keeps `go test ./...` green.

## Round 1 (reviewer on c6ad345, Opus, read-only): NOT READY

| # | Sev | Finding | Resolution |
| --- | --- | --- | --- |
| 1 | high | `pull_request` dry run builds binaries on the draft PR; contradicts "no workflow run" and AGENTS.md distribution-build rule | `pull_request` trigger removed; dry run is owner `workflow_dispatch` only; open decision 7 |
| 2 | med | Environment gate assumed; re-run by non-owner reuses approval | Publish re-checks actor; tag is the approval, protection optional (D8) |
| 3 | med | Release-exists check misses drafts | Publish lists all releases and fails on the tag name |
| 4 | med | Mutation list incomplete | FR-8 and the mutation table extended; `${{` in `run:` forbidden |
| 5 | med | Check-run query underspecified | Query by workflow file, event=push, success; in-progress fails with a wait message |
| 6 | med | Proxy/sumdb risk understated | Blocking open decision 2, recommends v0.1.1; AGENTS.md line says so |
| 7 | med | Archive modes not produced by flags | Staging `chmod`, `--format=ustar`, tests assert modes and owner |
| 8 | low | paths filter | Moot: trigger removed |
| 9 | low | macOS checksum tool | `shasum -a 256 -c` fallback in smoke script |
| 10 | low | "No tool is downloaded" | Reworded (setup-go installs Go) |
| 11 | low | Standing-approval reading | Open decision 6; AGENTS.md states the reading |
| 12 | low | Implementation before READY | Order note above |
| 13 | low | Tag fetch | Forced `+refs/tags/$TAG:refs/tags/$TAG`, `--no-tags` |
