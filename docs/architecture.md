---
title: "Architecture"
description: "How the compiler pipeline and its packages fit together."
category: project
---
`cmd/smill` delegates process exit to `internal/cli.Run`, which can be tested with
in-memory stdout/stderr. The CLI selects explicit inputs and performs bounded
reads. It never walks recursively or resolves content URLs.

`internal/frontmatter` handles UTF-8, exact boundaries, a YAML node tree, duplicate
keys and JSON-safe conversion. `internal/validate` checks known metadata types.
`internal/diagnostic` indexes source bytes into line/rune positions and stable codes.

`internal/markdown` parses GFM using Goldmark, independently segments root headings,
validates the limited heading attribute grammar and converts the AST to our IR.
Goldmark types do not escape this package. `internal/model` owns the versioned
serializable representation, stable IDs, ordered sections/cards and literal code.

Three direct dependencies are justified: Goldmark handles GFM parsing, a YAML
node tree supports duplicate keys and source locations, and the user-selected
Cobra v1.10.2 provides command/help/flag parsing. Other production operations use
the Go standard library. Dependencies are
pinned in `go.mod`/`go.sum`; no sibling project package is imported.

Future renderers consume IR and must establish their own layout and escaping
contracts. The Hugo renderer consumes the semantic IR directly and exports a complete
project; it never reparses original Markdown. See [Hugo contracts](hugo.md).
The navi renderer (`internal/render/navi`, [navi export](format/navi-profile.md)) likewise consumes
the IR, preserving complete frontmatter in a versioned `.navi.json` companion
for each exported sheet: a card is exported only when it carries `.navi` and an explicit ID, and the
package can neither execute commands nor use the network.

Go 1.27.2 was selected after checking Go release support and sibling policy on
2026-10-13. BundleDock pins that toolchain; WorkHarbor tests it alongside 1.26.9.
See [provenance](provenance.md) for inspected revisions.

`internal/importer/reference` is a pure adapter returning canonical Markdown plus
a structured loss report. CLI owns bounded input and exclusive output creation.
The adapter preserves provenance, archives legacy behavior and revalidates against
the core. It never fetches or executes external content. See [imports](imports.md).
