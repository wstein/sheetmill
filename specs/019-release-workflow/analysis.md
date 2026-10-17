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

## Round 2 (reviewer on a76b95c, Opus, read-only; ran tests, shellcheck, actionlint v1.7.7): NOT READY

| # | Sev | Finding | Resolution |
| --- | --- | --- | --- |
| 1 | high | Default dry-run label `v0.0.0-dryrun` rejected by `release-package.sh`, so the dry run could not work | Package script accepts `vX.Y.Z[-label]`; `TestDryRunLabelPackagesAndSmokes` packages, smokes and checks the dispatch default |
| 2 | med | Mutation tests were substring checks; 11 weakenings undetected | New rules: `set -euo pipefail` first in multi-line scripts, `|| true`/`|| :` forbidden, `if` forbidden on all steps except the three tag-only checks, exact `cmp`, refspec, `-ge 1`, smoke artifact name and call; 14 added mutations. Still string-level, not shell-AST: stated limitation |
| 3 | low | Dispatch not owner-only; `v0.1.0` accepted as a dry label | Actor check runs on every run kind; dry label requires a suffix |
| 4 | low | Package script accepted a newline in the version | Newline rejected, test added |
| 5 | low | Stale status, unchecked tasks, FR-2 order | Status, checkboxes, FR-2 order aligned |
| 6 | low | Notes script error on a parentless commit | Guarded with `git rev-parse -q --verify`; `@user` escaping not done (subjects are the owner's own commits) |
| 7 | low | Dead loop in test | Removed |
| 8 | low | Tag could move between `ls-remote` and `gh release create` | Accepted (owner-only, documented) |

## T008 local verification (author, at the round 3 candidate)

Run: `gofmt -l`, `go vet ./...`, `go test ./...`, `go test -race ./internal/releaseworkflow`,
`make check`, `SHEETMILL_HUGO=/opt/bundledock/toolbox/bin/hugo make test-hugo`; `bash -n` on every
extracted `run:` block. All passed. Not run: any workflow, `make test-browser*` (Node is not
installed), actionlint (not installed; the round 2 reviewer ran v1.7.7 clean), macOS/arm64
execution of the scripts, the GitHub-side tag path.

## Round 3 (reviewer on b0c604a, Opus, read-only; ran tests, actionlint 1.7.7, a manual dry run through packaging and smoke): NOT READY

The reviewer found no high findings and confirmed rounds 1 and 2 resolutions, the dry run by
hand, and the authorization wording. 33 new mutations: 6 caught, 10 harmless, 17 slipped.

| # | Sev | Finding | Resolution |
| --- | --- | --- | --- |
| 1 | med | Gates disarmable by `exit 0`, `\|\| echo`, `set +e`; release-exists `case` loose | Assertion lines must match `test ... \|\| { echo ... >&2; exit 1; }`; `exit 0`, `set +e`, `\|\| echo`, `\|\| exit` forbidden in scripts; release-exists branches pinned; 6 mutations added |
| 2 | med | ACTOR/OWNER env could be made always equal | Exact env of both owner checks pinned; mutations added |
| 3 | low | Spec/plan claimed a shell parser | Claim removed; limitation listed in the spec |
| 4 | low | tasks/plan drift | Aligned |
| 5 | low | Standing approval is the spec's reading | Open decision 6 (owner confirms before the second release) |
| 6 | low | HEAD check and publish checkout ref untested | Pinned, mutations added |

Verdict: three rounds are used. No fourth review was run on the final fixes; the author's
re-run of the workflow tests passes. The 17 slips were not all re-tested individually.
Remaining known gap: tests are string/line-level (spec Known limitations). Treat the
verdict as NOT READY pending the owner's review of the workflow PR.

## Post-round-3 review (coordinator's independent Opus review of 2b4a92f): CLEAR, 4 medium, 11 low

Runtime scenarios behaved correctly. 48 fresh mutants: 21 killed, 27 survived, 10 real gaps. Addressed
in the next commits: F1/F2/F3/F5/F6/F9 first by `testdata/release.golden.json` (the whole workflow,
comments dropped; a change detector only), plus a runtime
`git rev-parse HEAD` = verified commit assertion in build and smoke and an exact-env check;
F4/F14 by rewriting docs, spec and AGENTS.md (default `v0.1.1`, retag of `v0.1.0` discouraged
with the effects stated as fact, standing approval marked an open owner confirmation); F7/F8 by
script tests (failing `check`, extra archive member, sorted member order); F10/F12 documented;
F11 by also requiring `github.actor` to be the owner; F13 by the check.yml wording; F15 by the
harness error text "HARNESS ERROR, NOT A KILL". Limitation: the golden makes any workflow edit
a deliberate, reviewed change; it does not judge whether an edit is safe.

## Confirmation review of 7b63e85 (coordinator, PR #49): CLEAR, 2 medium, 8 low

The golden alone left 9 mutants alive once regenerated, so those items were only *detected as a
change*, not prevented. Now also covered by semantic rules independent of the golden
(`semantic()` in `workflow_test.go`): no assignment to ACTOR/OWNER/PUSHER/COMMIT/TAG/REPO and
friends in any script, `n`/`remote`/`existing`/`commit` assigned once from a command
substitution, no eval/export/source, only `https://github.com/$REPO` URLs, exact owner,
pusher, `-eq 3`, sorted-checksum, `commit=$GITHUB_SHA`, check-run and second-build lines, no
job env, no env on gates, per-action `with` allow-lists, no `cp`/`mv` between builds.
The golden is no longer part of `violations()`; it is its own test, and updating it always fails
the run and is refused in CI. Script tests added: dates from SOURCE_DATE_EPOCH, other umask and
file times, failing `--help`, non-executable binary, baseline-tag choice (older stable tag,
newer pre-release). Docs: retag hazard applies to any re-created tag (re-run or use the next
patch), expired artifacts, golden churn, org-owner limitation. Known survivor: a change of the
package call arguments (`"${VERSION}x"`) is caught only by the golden.

## F11 studio flake (outside 019)

`go test ./internal/studio -run TestInPlaceWritesNeverAcceptMixedSnapshot -count=150 -cpu 1,2`
failed reproducibly before the fix ("accepted torn in-place read"). Cause: the test's writer can
be descheduled mid-write, leaving a file that is torn at rest; two identical torn reads are then
correct behavior of the stability check and no reader can tell them from a stable file. The
test now counts writer progress and asserts only for reads during which the writer progressed.
`-count=300 -cpu 1,2` passes after the change. Not verified: that the reworked test still fails
when the stability check is removed (the attempted mutation did not compile).
