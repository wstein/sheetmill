---
title: WorkHarbor bootstrap
category: reference
description: Notes on the WorkHarbor bootstrap repository workflow as pinned, which is a development scaffold with no implemented whr product CLI.
tags: [workharbor, whr, bootstrap]
cheatsheet:
  updated: "2026-10-16"
source:
  repository: https://github.com/wstein/workharbor
  revision: 916add9a7c46dc2d8d9166a01947be9495789cc5
  path: README.md
  license: EUPL-1.2
  attribution: SheetMill contributors
  note: Independently authored for the user-selected bootstrap repository; no whr product CLI is implemented at this revision. Examples are inert.
---
This reference documents the WorkHarbor bootstrap repository. At the pinned revision it is a Spec Kit project, not a working platform, and it has no implemented `whr` product CLI. SheetMill displays the development workflow without executing it.

## Repository workflow {#workflow}
### Start the WorkHarbor checkout {.col-span-2 #start-checkout}
```sh
git clone https://github.com/wstein/workharbor.git workharbor
cd workharbor
make check
```

The repository requires Git, Make, Python 3.11+ and Bash. Repository checks validate development scaffolding; passing them does not provision a workspace or start a supervisor.

### Select the lifecycle feature {#select-feature}
```sh
python3 scripts/select_feature.py 001-workspace-lifecycle
```

Select an existing direct `specs/` directory containing `spec.md`. The script writes ignored checkout-local `.specify/feature.json`; it does not implement the selected feature.

### Run repository tests {#repository-tests}
```sh
make test
```

The target runs Python unittest discovery under `tests/`. `make check` uses the repository's isolated check wrapper. These are contributor operations, not a `whr` command interface.

## Design boundaries {#boundaries}
### Distinguish plans from a product {.col-span-2 #implementation-status}
The README describes workspace lifecycle implementation as pending. Phoenix/Elixir/LiveView, a Go agent and initial SQLite metadata storage are architectural choices, not evidence of an installed platform.

Do not borrow `whr` commands from the archived `wstein/workharbor-old` repository for this source revision. The corpus ID and filename `whr` identify this reference input; they do not claim a working executable.

### Inspect the workspace example {#workspace-example}
The `examples/workspace` directory is a layout example, not a provisionable fixture. A real workspace needs its own project repository and reviewed image pins; the example intentionally supplies no fake digest.

Read the repository's guidance before adding real project submodules. SheetMill does not create repositories, fetch projects or activate containers from this sheet.

### Keep tool bundles independent {#tool-boundary}
WorkHarbor's proposed workspace configuration keeps optional bundles separate from language images. No implemented BundleDock integration is claimed by this bootstrap.

See [BundleDock volume options](bdock.md#volume-help) for the independently implemented Docker tool-volume interface. This reference link connects the concepts without running either project's workflow.
