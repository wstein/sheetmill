---
title: BundleDock CLI
category: reference
description: Discover the bdock CLI, initialize shared tool volumes, pin a reviewed image, and upgrade the source catalog lock.
tags: [bundledock, bdock, oci]
cheatsheet:
  updated: "2026-10-16"
source:
  repository: https://github.com/wstein/bundledock
  revision: 81c2688ea8dcb782f34b51f96c3c2f18bdd00d85
  path: cmd/bdock/tools.go
  license: EUPL-1.2
  attribution: SheetMill contributors
  note: Independently authored from the pinned command definitions; examples are inert.
---
BundleDock shares pinned tools through OCI volume carriers. These examples describe the verified CLI source; SheetMill does not run Docker or contact a registry.

## Command discovery {#discovery}
### Discover BundleDock commands {.col-span-2 #discover-commands}
```sh
bdock --help
bdock tools --help
bdock --version
```

The root command offers `tools`; running either group without a subcommand prints help. `--version` reports the build's version.

### Inspect volume options {#volume-help}
```sh
bdock tools volume --help
```

Read the image and bundle options before initializing a shared volume. Help is descriptive; the volume operation itself requires Docker.

## Shared tool volumes {#volumes}
### Initialize a tool volume {.col-span-2 #initialize-volume}
```sh
bdock tools volume --image bundledock-toolbox:poc --bundle toolbox
```

This is the CLI's source default image, not a claim that this tag is published or suitable for production. Choose a reviewed immutable image reference for actual use. The operation initializes and verifies an image-derived Docker volume.

### Choose a bundle identity {#bundle-options}
| Flag | Source default | Meaning |
| :-- | :-- | :-- |
| `--image` | `bundledock-toolbox:poc` | OCI tool bundle image |
| `--bundle` | `toolbox` | Bundle name in the image-derived volume namespace |

The command accepts no positional arguments. Keep consumer credentials and writable configuration outside shared read-only tool storage.

## Catalog maintenance {#maintenance}
### Upgrade the source catalog {.col-span-2 #upgrade-catalog}
```sh
bdock tools upgrade --directory bundles/toolbox
bdock tools upgrade ripgrep --directory bundles/toolbox
```

Resolve the newest stable versions allowed by `packages.txt` into `packages.lock.json`. Omit the optional tool to consider the catalog; name one tool to restrict the update. This operation can use the network and write the lockfile when run by a person.

### Keep references reproducible {#reference-boundary}
Pin the source revision when documenting flags. Bundle publication and consumer activation are separate operations from these CLI commands.

For documenting the verified command inventory as a reference, see [SheetMill semantic inspection](smill.md#inspect-ir). The link and code blocks remain inert content in this library.

## Reproducible selection {#reproducibility}
### Pin a reviewed image {.col-span-2 #pin-reviewed-image}
```sh
bdock tools volume --image quay.io/bundledock/toolbox@sha256:40bae37ef7fd7b8e7131803850801ed41cfb23a6ef774bc2e40f95ee81e59a1c --bundle toolbox
```

The pinned repository README records this index for toolbox `2026.10.10-b4.1`. It is a source-recorded publication receipt, not a claim about today's registry availability or the latest catalog. Review the receipt and consumer requirements before real use; SheetMill does not pull or initialize it.
