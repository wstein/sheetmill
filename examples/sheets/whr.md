---
title: WorkHarbor CLI
category: reference
tags: [workharbor, whr, agents]
source:
  repository: https://github.com/wstein/workharbor
  revision: 5d6a8a423ec509383e1cf169bdea0a76d58ecd92
  path: internal/cli/commands.go
  license: EUPL-1.2
  attribution: SheetMill contributors
  note: Independently authored from the implemented whr CLI, not the workharbor-nxt bootstrap; examples are inert.
---
WorkHarbor supervises agent tasks and exposes human decisions through `whr`. These references describe a pinned implementation; SheetMill starts no agent, service or account session.

## Inspect the work {#inspection}
### List WorkHarbor tasks {.col-span-2 #list-tasks}
```sh
whr ls
whr ls --all
whr --json ls
```

The default list contains unfinished tasks; `--all` includes finished tasks. `--json` selects the API envelope instead of the text table. These commands require your configured supervisor when actually run.

### Discover CLI help {#discover-help}
```sh
whr --help
whr run --help
whr version --json
```

Help and version do not need the API configuration. The source defines `--config` for an explicit configuration path; otherwise it uses `WHR_CONFIG` or the default home configuration.

### Follow task events {#follow-events}
```sh
whr logs 'my-repo#42' -f --since 100
```

`-f` follows events and reconnects; `--since` starts after the specified event ID. A task reference may be a full ID, a unique ID prefix or an issue such as `my-repo#42`. Ambiguous references need disambiguation.

## Tasks and decisions {#decisions}
### Start an issue task {.col-span-2 #start-task}
```sh
whr run https://github.com/example/project/issues/42 --agent demo/builder
```

Supply one issue URL and the required `--agent <workspace>/<role>`. Replace the illustrative project and agent with your configured identities before use. Running this command creates work; it is only displayed here.

### Inspect pending decisions {#pending-decisions}
```sh
whr inbox
whr approve --help
whr reject --help
```

The inbox lists open decisions. Approval takes a decision reference; review approvals also need the exact commit shown through `--sha`. Rejection accepts an optional `--reason`. Reading this sheet grants no approval.

### Keep tool setup separate {#tool-boundary}
WorkHarbor's `whr tools build` and BundleDock's OCI tooling are separate interfaces. Do not substitute one command tree for the other.

See [BundleDock volume options](bdock.md#volume-help) for its Docker-backed tool-volume command. SheetMill verifies this link while treating both products' snippets as inert content.
