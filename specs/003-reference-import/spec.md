# 003 — Controlled Reference import

Created: 2026-10-13 · Status: implemented, review pending · Depends on: 002

## User scenarios and acceptance

1. P1: A user converts one explicitly selected local Fechin/reference Markdown
   file into canonical GFM+ with repository, path, revision, license, attribution
   and source SHA-256 provenance. Markdown body bytes stay unchanged.
2. P1: A user previews deterministic JSON loss reports without writing anything.
   Unsupported structural syntax fails. Plugin/theme behavior is recorded as inert
   source metadata and reported, never executed or interpreted as renderer policy.
3. P1: A user reruns conversion after editing a target. Existing target files,
   symlinks and directories are refused; no force/overwrite mode exists.

## Requirements

- FR-001: Only the explicitly selected Reference adapter; no auto-detection guessing.
- FR-002: Explicit provenance flags; full lowercase 40/64-hex revision; no clocks.
- FR-003: Preserve source content, original metadata and attribution; source fields
   do not auto-approve redistribution. Conflicting existing provenance is an error.
- FR-004: Move legacy plugin/theme configuration into inert provenance metadata;
   report changes with source-positioned stable codes and severity.
- FR-005: Validate converted bytes with the canonical parser before any output write.
- FR-006: Deterministic Markdown conversion and schema-1 JSON report, with source
   and converted hashes. Dry-run does not write. Exclusive output creation.
- FR-007: No fetching or code/plugin execution; no implicit import crawling.
- FR-008: Test the pinned actual Python source from scratch storage without bundling
   upstream GPL content. Bundle only independently authored conversion fixtures.

## Success criteria

Conversion reparses successfully; second conversion of identical source/provenance
is byte-identical. Disabled plugins and theme hints appear in the report. Unsupported
attrs prevent writes. Existing output remains byte-identical after a refused rerun.
The actual Python benchmark imports with card/fence/table structure retained.

Out of scope: TLDR/navi/Devhints adapters, bulk imports, source retrieval, renderer
execution, publication and deciding legal permission to republish.
