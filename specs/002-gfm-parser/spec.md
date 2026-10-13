# 002 — GFM+ parser and semantic IR

Created: 2026-10-13 · Status: implemented, review pending · Depends on: 001

## User scenarios and acceptance

1. P1: An author checks a Markdown sheet and receives actionable errors. Duplicate
   YAML keys, missing titles, malformed attrs and orphan cards report file/line/column.
2. P1: A tool inspects a valid sheet as deterministic JSON retaining ordered
   sections, cards, GFM blocks, inline semantics, layout hints and source spans.
3. P2: An author can select files/one-level directories without implicit crawling.
   Nonregular files and symlinks fail; source code is never executed.

## Requirements

- FR-001: Implement `docs/format/gfm-plus.md` exactly; preserve unknown metadata.
- FR-002: Strict YAML node validation and stable diagnostic codes.
- FR-003: Root ##/### segmentation; nested headings/fence literals stay content.
- FR-004: Preserve tables/alignment, lists/tasks, links, quotes, code and formatting.
- FR-005: Stable IDs, schema-1 JSON, no clocks or absolute paths in IR.
- FR-006: Bounded explicit input discovery, stderr diagnostics, documented exit codes.
- FR-007: Independent synthetic fixtures, checked-in goldens and fuzz targets.

## Success criteria

All valid fixtures parse; invalid fixtures fail at meaningful source positions.
Repeated output equals golden bytes. Nested structure and literal fence tests pass.
Race tests and vet pass; bounded fuzzing records no crashes. No renderers exist.
