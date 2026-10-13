# Verification record — 2026-10-13

Implementation reviewed locally on `codex/001-core-foundation` in
`~/github.com/wstein/sheetmill`, Go 1.27.2, macOS arm64. Existing editor settings
were preserved. No distributable binary, renderer, PDF, site, release or deployment
was produced. Tests compile their own temporary test executables.

## Actual commands and results

Local Go commands used `GOPROXY=off` and
`GOCACHE=/private/tmp/sheetmill-go-cache`; the two pinned dependencies were already
cached. Ordinary CI downloads pinned modules as development dependencies.

| Command | Result |
| --- | --- |
| `gofmt -w cmd internal`, then empty `gofmt -l cmd internal` | Pass |
| `go test ./...` | Pass |
| `go test -race ./...` | Pass, including final CLI/CRLF/nested-fence regressions |
| `go vet ./...` | Pass |
| `go mod verify` | `all modules verified` |
| `go test ./internal/frontmatter -run '^$' -fuzz FuzzParse -fuzztime 5s -parallel 2` | Pass; final run 39,266 executions |
| `go test ./internal/markdown -run '^$' -fuzz FuzzParse -fuzztime 5s -parallel 2` | Pass; final run 41,207 executions |
| `go test ./internal/importer/reference -run '^$' -fuzz FuzzConvert -fuzztime 5s -parallel 2` | Pass; 29,580 executions |
| `SHEETMILL_REFERENCE_PYTHON=/private/tmp/sheetmill-reference-python.md go test ./internal/importer/reference -run TestPinnedPython -v` | Pass; verified input hash, 17 sections, 119 cards, 2 inert-metadata findings |
| `git diff --check` and staged diff checks per commit | Pass |
| Spec Kit `check-prerequisites.sh --json --require-spec --require-tasks --include-tasks` with explicit feature directory | Pass for 001, 002, 003 |

Golden JSON was regenerated explicitly during initial implementation using
`go test ./internal/markdown -run TestGolden -update` and
`go test ./internal/importer/reference -run TestGolden -update`, inspected for
structure/provenance/literal fidelity, then checked without update flags.

Tests were added before implementation and observed failing for missing parser/CLI
entrypoints and later for missing import conversion/CLI/loss-report behavior.
Tests also caught Goldmark's split text nodes in attributes; that was corrected.
Regression coverage includes duplicate keys, invalid metadata/UTF-8, conflicting
attrs, code literals, nested headings, CRLF, deterministic IDs/JSON/conversion,
dry-run, error JSON, refusal to overwrite, ancestor/output symlinks and stdout failure.

A source inspection of production Go packages found no `os/exec`, HTTP client,
`syscall`, `unsafe` or dynamically loaded plugin imports. The Reference converter
contains no filesystem or network operations. This source check and bounded fuzzing
do not replace a dependency vulnerability scan or an independent security review.

## GitHub controls and remaining review

Repository is public with approved description/topics. Main protection was set and
read back: required `core` check, up-to-date branch, one approving review, stale-review
dismissal, admin enforcement, linear history, conversation resolution, no force pushes
or deletion. Automatic merges are disabled. CI has only contents-read permissions
and SHA-pinned Actions; it publishes no artifacts.

[Draft PR #1](https://github.com/wstein/sheetmill/pull/1) was opened and attached to
the chat. Both push and pull-request `core` checks passed for implementation
revision `2f3f8423224648d45c0c58ad747406c8d8757c5b`. Final checklist-only updates
are also checked by CI; the latest result is reported in the chat.
Independent review is pending and remains a merge gate. Tests and fuzzing are local
author evidence, not independent approval. Optional upstream content stays outside
the repository; ordinary CI does not fetch it. No dependency vulnerability scan,
full external-format compatibility claim, logo raster/monochrome QA or renderer
validation is asserted.
