# 004 — Hugo renderer

Created: 2026-10-13 · Status: draft, specification only · Depends on: 003-reference-import

## User scenarios and acceptance

P1: An author can produce responsive static cards, accessible navigation and search metadata from the canonical IR.
Given identical source, configuration and pinned tools, output is deterministic.
Unsupported content produces actionable diagnostics rather than disappearing.

## Requirements

- FR-001: Consume canonical IR; never parse external source formats directly.
- FR-002: Define escaping, layout and output-version contracts before implementation.
- FR-003: Require explicit output destinations; no implicit network/tool downloads.
- FR-004: Test semantics without running source snippets or dynamic suggestions.
- FR-005: Retain source attribution and license metadata in generated references.
- FR-006: Obtain separate implementation/build/publication authorization.

## Success criteria

After authorization: semantic fixtures and deterministic output tests pass; a human
reviews generated layout/behavior. Tool Hugo is explicitly pinned and its fonts,
themes, dependencies and licenses are documented. No build or publication is
performed under the current milestone.

## Open decisions

Search implementation and theme ownership; no sibling CSS copying.
