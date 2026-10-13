# Additional reference corpus

T042–T046 extend the same GFM → CLI → Hugo → browser harness. The default
selection remains `smill`; `SHEETMILL_E2E_CORPUS=all` selects the three explicit
references. Sibling CLIs are neither installed nor invoked. Source inspection is
read-only; sheets are independently authored by SheetMill contributors under
EUPL-1.2, rather than copied upstream cheatsheets.

## Verified command inventory

BundleDock source: [wstein/bundledock at 81c2688ea8dcb782f34b51f96c3c2f18bdd00d85](https://github.com/wstein/bundledock/tree/81c2688ea8dcb782f34b51f96c3c2f18bdd00d85).
Inspected `cmd/bdock/root.go`, `cmd/bdock/tools.go`, `README.md` and `LICENSE`.

| Displayed command | Verified contract |
| --- | --- |
| `bdock --help`, `bdock tools --help` | Cobra help; root/tools groups print help without arguments |
| `bdock --version` | Root version flag |
| `bdock tools volume --help` | Cobra help for the volume command |
| `bdock tools volume --image … --bundle …` | No positional args; defaults `bundledock-toolbox:poc` / `toolbox`; Docker operation |
| `bdock tools upgrade [tool] --directory …` | At most one tool; resolves catalog into lockfile; may fetch/write |

The example's source-default image is explicitly not a publication receipt. No
bundle image was pulled, no volume was initialized and no catalog was upgraded.

WorkHarbor source: [wstein/workharbor at 5d6a8a423ec509383e1cf169bdea0a76d58ecd92](https://github.com/wstein/workharbor/tree/5d6a8a423ec509383e1cf169bdea0a76d58ecd92).
Inspected `internal/cli/root.go`, `internal/cli/commands.go`, `cmd/whr/main.go`,
`cmd/whr/tools.go` and `LICENSE`.

The similarly named local checkout points to `wstein/workharbor-nxt` at
`916add9a7c46dc2d8d9166a01947be9495789cc5`. Its README explicitly describes a
bootstrap with product implementation pending, and it has no implemented `whr`
command tree. This sheet therefore documents the existing `wstein/workharbor`
implementation. No sibling checkout was modified and no sibling CLI was run.

| Displayed command | Verified contract |
| --- | --- |
| `whr ls`, `whr ls --all`, `whr --json ls` | Unfinished/all task selection; global JSON API-envelope output |
| `whr --help`, `whr run --help` | Cobra help without API initialization |
| `whr version --json` | Binary's version passthrough, no API configuration |
| `whr logs 'my-repo#42' -f --since 100` | Exactly one task; follow/reconnect; event-ID offset |
| `whr run <issue-url> --agent demo/builder` | Exactly one issue URL; required workspace/role agent |
| `whr inbox`, `whr approve --help`, `whr reject --help` | List decisions; help for decision actions; review approval requires shown SHA |

Names/URLs in the task example are illustrative, not configured accounts.
`whr tools build` is mentioned to distinguish its interface, not offered as a
BundleDock alias. Source flags, argument validation and behavior were read
statically; no command was tested against a live supervisor.

## Actual verification

On 2026-10-13, from `/private/tmp/sheetmill-corpus-20261013`, using Node
24.21.0, Playwright 1.64.0 / Chromium 156.0.8078.4, Go 1.27.2 and Hugo 0.167.0:

| Check | Result |
| --- | --- |
| `npm run test:harness` | 15 passed, including invalid selection and cross-card/permalink oracles |
| Default `npm test` | 6 passed; the single-sheet MVP remains the default |
| `SHEETMILL_E2E_CORPUS=all npm test` | 12 passed |
| Repeated all-corpus suite with fresh owned roots | 12 passed again |
| `go test -race ./...`, `go vet ./...`, `go mod verify` | Passed, including installed-Hugo integration |
| `gofmt -l cmd internal assets`, `git diff --check` | Empty/passed |

The test PATH was `/private/tmp/sheetmill-node24/bin:/opt/homebrew/bin:/usr/bin:/bin`.
An explicit lookup check confirmed `bdock` and `whr` were absent from it (the
user's unrelated `~/.local/bin/whr` was excluded). The harness neither installs
nor invokes either CLI. Fixed subprocess stages are Go version/build, temporary
SheetMill version/check/inspect/export, and installed Hugo version/build. Browser
clipboard contents are read back as text, never pasted into a command runner.

Both scripted viewports exercised each product's complete shared journey and
exact multiline copy oracle, plus `bdock → smill#inspect-ir` and
`whr → bdock#volume-help`. The no-JavaScript project also followed those links.
Every build compared two complete public file maps byte-for-byte, every selected
source hash stayed unchanged, and every browser context enforced zero off-origin
requests. The source inputs contain 20 cards across three sheets.

Baseline three-sheet snapshots contain 16 public files; sorted relative filename
+ NUL + bytes + NUL SHA-256 is
`1c8580d5cc449e01b1ef71ac66137040a933e80947be16188bd74492f4814e0c`.
Temporary inert variants deliberately differ and are compared with their own
independent second build. Original source hashes:

| Input | SHA-256 |
| --- | --- |
| `smill.md` | `4d3228787a9570ac1cdb85207e458f590c70255c57b3680006d4e9ff99310f27` |
| `bdock.md` | `e83ebcf1513616527a5f65c1292ea36ed3c383dd7ba2d4f3e3a9cbce8aaebaaf` |
| `whr.md` | `fcf3d7a600e09ab7faefde476ddf993d1ae5c9425736c77977ad536031543adc` |

Local evidence is retained under `.local/e2e/runs/`. Author-local results are not
hosted Linux CI results. CI now runs both default and all-corpus suites, without
artifact uploads. No sibling source, product runtime, compiler contract or public
hosting configuration changed. Existing main-only Pages publication remains a
separate workflow; merging these example files would include them in its library.

## Review

Independent Sol review is pending; findings and exact reviewed revision will be
recorded before the PR handoff.
