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
