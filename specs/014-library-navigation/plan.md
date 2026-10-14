# Implementation Plan: Library navigation

**Branch**: `codex/014-navigation-complete` | **Date**: 2026-10-14 | **Spec**: [spec.md](spec.md)

**Input**: Settled first-slice specification in `specs/014-library-navigation/spec.md`.
**State**: Implemented under the user’s 2026-10-14 handoff; see tasks and
quickstart for verification. No deployment.

## Summary

Add optional per-sheet navigation group/order with one shared validator and one
shelf comparator. Preserve semantic categories, main page route order, related
relevance, URLs and generic metadata. Archive upstream Reference navigation before
validation and prove navi companion preservation without command changes.

## Technical Context

**Language/Version**: Existing Go version from `.go-version` (1.27.2).
**Primary Dependencies**: Existing YAML/Goldmark/Cobra; installed Hugo 0.167.0 and
existing pinned Playwright harness for local verification. No new dependencies.
**Storage**: Existing source frontmatter/semantic metadata and generated project;
no new project config/manifest or renderer filesystem reads.
**Testing**: Go negative tables, parser/import goldens, renderer membership fixtures,
byte determinism, navi pair comparisons, existing browser harness.
**Target Platform**: Existing CLI platforms and static browser output.
**Project Type**: Go compiler/CLI with static website rendering.
**Performance Goals**: Preserve existing performance checks; group-member sorting
is O(n log n), with no discovery/network work or extra runtime dependency.
**Constraints**: Closed namespace; no defaults in IR; no schema bump; no execution,
PDF, deployment, merge or reordering real libraries in this increment.
**Scale/Scope**: Two optional fields; fixtures cover ordered/unordered/tied/overridden/
uncategorized groups, direct IR, imports and metadata companions.

## Constitution Check

Pre-research, post-design and implementation: PASS within recorded scope.

- Determinism: explicit total comparison; stable source/card/URL identity.
- Canonical source: public format/schema/diagnostic changes planned with tests;
  source compatibility restriction documented despite unchanged IR version.
- Loss-aware imports: early inert archival preserves arbitrary upstream values.
- Inert input: no snippets, navi, dynamic providers or source resources execute.
- Focused design: existing dependencies and shared functions; no manifest/config I/O.
- Review/publication: implementation authorization is recorded in AGENTS.md;
  exact-revision independent review is required before merge.

## Project Structure

### Documentation (this feature)

```text
specs/014-library-navigation/
  spec.md
  plan.md
  research.md
  data-model.md
  contracts/navigation.md
  quickstart.md
  tasks.md
  checklists/requirements.md
```

### Source Code (future implementation)

```text
internal/validate/navigation.go             shared focused validator
internal/validate/navigation_test.go        scalar/shape/boundary tests
internal/validate/validate.go               delegation from Metadata
internal/frontmatter/frontmatter_test.go    source-spanned negative cases
internal/render/hugo/navigation.go          derived settings/comparator integration
internal/render/hugo/hugo.go                direct-IR boundary and shelves() copies
internal/render/hugo/navigation_test.go     multi-sheet/permutation/related tests
internal/importer/reference/reference.go    early single-key archival
internal/importer/reference/reference_test.go
internal/render/navi/navi_test.go           preservation/invariance regression only
testdata/reference-compatible/navigation.md[.json]
testdata/hugo/navigation/                   independent multi-sheet fixtures
schema/cheatsheet.schema.json
docs/format/gfm-plus.md, docs/hugo.md, docs/imports.md, docs/diagnostics.md
tests/e2e/navigation.spec.mjs
```

**Structure Decision**: Preserve generic `model.Document.Metadata`. An unexported
or JSON-excluded navigation setting on Hugo's internal page holds presence/order;
page.Category remains original. `shelves()` derives presentation groups and sorts
copied member slices. Its existing group-label data continues to drive both home
layouts and the sidebar. No new source-facing CLI flags or file-loading options.

## Phases

0. Research: code claims verified in [research.md](research.md); no unresolved
   technical choices remain.
1. Design: [data-model.md](data-model.md), [contract](contracts/navigation.md) and
   [quickstart](quickstart.md) define interfaces, edge cases and future verification.
2. Implementation (future): execute [tasks.md](tasks.md) after authorization;
   failing tests precede rules, rendering and import changes.
3. Verification/handoff (future): source checks, all unchanged golden bytes, local
   browser journeys, CI and exact-revision independent review. No deployment.

## Complexity Tracking

No constitution exception or extra abstraction beyond a shared validation function
and a deterministic comparator is required. Adding a manifest or general metadata
warning pipeline is explicitly excluded.
