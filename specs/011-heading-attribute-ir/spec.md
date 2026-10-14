# 011 — Structural heading titles without attribute syntax in the IR

Created: 2026-10-13 · Status: draft, specification only · Depends on: 002-gfm-parser, 004-hugo-renderer
Authorization: 011 implementation is not approved; other implemented features
do not authorize an IR heading-contract change.

## Problem

`internal/markdown/markdown.go` `heading()` parses the trailing `{.class #id}` group
of section/card headings, joining only `Text` segments that are contiguous in the
source. `internal/render/hugo/hugo.go` `renderer.heading()` then strips the suffix
again from the IR heading's trailing `Text` children, joining all trailing `Text`
nodes. Two copies of one rule can disagree: literal attribute syntax can be rendered,
or title text can be truncated. By contract (docs/format/gfm-plus.md, "Heading
attributes") the IR `heading` keeps the original inline nodes for source fidelity and
`title` is a plain string, so no IR field carries title inline nodes without the group.
Listed in docs/recommendations.md as a follow-up needing its own reviewed PR.

## User scenarios and acceptance

1. P1: An author writes `## Pipes {#pipes}` or `### *Fast* `cut` {.tip #cut}`. Every
   renderer shows the title without the `{...}` group, with emphasis, code and
   links intact, and without re-deriving the rule.
2. P1: An author writes `## Escaped \{#x}` or a title containing an entity or escape
   before a group. The escaped braces stay literal; the group is recognized exactly
   when `markdown.heading()` recognizes it, in every renderer.
3. P1: Malformed input keeps its diagnostics: two groups give `ATTR001`; a
   duplicate/conflicting key gives `ATTR002`; unsupported tokens give `ATTR001`.
   Spans are unchanged.
4. P2: A future Studio editor can load a heading, edit its title and preserve its
   attributes, because `heading` keeps the original nodes while `title_nodes`,
   `id`, `classes` and layout carry the parsed result separately.

## Requirements

- FR-001: One rule, one implementation: the parser alone decides where the
  attribute group starts. No renderer, importer or tool re-parses heading text.
- FR-002: Sections and cards gain parser-produced title inline nodes (working name
  `title_nodes`) that exclude the attribute group and the whitespace before it.
  `heading` keeps the original inline nodes unchanged, preserving the documented
  source-fidelity contract; `title` stays the plain decoded string.
- FR-003: Escaped braces (`\{#x}`, `\\{#x}` where the count of backslashes decides)
  remain literal text. Emphasis, strong, code spans, links and images in titles are
  preserved as nodes.
- FR-004: Remaining nodes keep exact half-open byte spans in the entire original
  input. The removed suffix has its own span (already used for ATTR diagnostics).
- FR-005: The group must be recognized across non-contiguous `Text` segments
  (entity, escape or soft-break boundaries) identically to today's parser rule.
  An attribute-like `{...}` after a code span, link or other non-`Text` node as the
  last child is not a group and stays literal; the title is unchanged.
- FR-006: `renderer.heading()` in internal/render/hugo is deleted or reduced to
  rendering the IR children; Typst and navi renderers (005, 006) rely on the same
  guarantee. A test fails if a renderer imports heading-attribute parsing.
- FR-007: Schema decision (open decision 1) is recorded in docs/format/gfm-plus.md
  (IR `schema: 1` section), `schema/cheatsheet.schema.json`, docs/hugo.md (which
  states the renderer consumes schema-1 IR) and docs/architecture.md. A schema bump
  updates `model.Schema`, the JSON Schema, the compatibility notes
  (docs/format/compatibility.md) and rejects unknown versions with a diagnostic.
- FR-008: Reference import re-parses its input and output with `markdown.Parse`
  (internal/importer/reference/reference.go), so it inherits the change; its
  canonical Markdown output and fixtures must stay byte-identical.
- FR-009: Hugo output is byte-identical for existing fixtures, except cases where
  the old duplicated rule was wrong; each such diff is listed and justified.
  `data/ir/<sheet>.json` gains the new field and stays deterministic.
- FR-010: No new diagnostic code unless justified. Existing `ATTR001`/`ATTR002` and
  section-heading `#id`-only behavior are unchanged and documented in docs/cli.md.

## Edge and negative cases (table-driven tests)

| Input heading | Expected |
| --- | --- |
| `## T {#a} {#b}` | `ATTR001`, one trailing group only |
| `## T {#a #b}` | `ATTR002`, duplicate id |
| `## T \{#x}` | no group; literal `{#x}` in title |
| `## T &amp; {#x}` | group recognized; title `T &` with entity preserved |
| `## T \* {#x}` | group recognized across escape boundary |
| `` ## `c` {#x} `` | title is the code span; id `x` |
| `## [l](u) {#x}` | link kept; id `x` |
| `` ## T `c`{#x} `` | decided: group recognized after code span only if the rule is kept; otherwise literal (record in test) |
| `## T {.bogus}` | `ATTR001`, unsupported token |
| `## {#x}` | empty title behavior unchanged and tested |

## Success criteria

Failing tests are written first. Golden JSON is regenerated only with
`go test ./internal/markdown -run TestGolden -update` and semantically inspected
(only the new title-node field, plus `schema` if bumped, changes). A fuzz target asserts that for any
heading the parser's title nodes contain no recognizable attribute group and that
spans remain within the input. Hugo fixtures compare byte-for-byte before/after.
`gofmt`, `go test ./...`, `go vet ./...` and `git diff --check` pass. Independent
review and human approval are required before merge.

## Open decisions

1. Additive field within `schema: 1` (existing consumers keep working; `heading`
   is untouched) versus a `schema: 2` bump that makes `title_nodes` required.
2. Field shape: a node slice versus a single cut offset/span into `heading` that
   renderers apply without interpreting syntax.
3. Whether attribute groups after non-`Text` last children remain literal (FR-005).
