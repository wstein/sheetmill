# Plan — 005

Specification only. Start implementation after separate authorization and after
002-gfm-parser is accepted. Build a renderer around canonical IR, maintaining portable
semantic metadata and renderer-owned layout. Pin the external tool before use.

1. Close open decisions in spec.md (page defaults, font profile, accessibility
   target, Typst version) and record them as a short contract document.
2. Add licensed TTF/OTF font assets with provenance and hashes under `assets/`
   (separate approval; no sibling-repository copying).
3. Write fixtures and failing tests for A1–A12: escaping table tests, forbidden-
   construct scan, golden `.typ` bytes, order/host independence, fuzz target.
4. Implement `internal/typst` (escape, page model, attribution page) and
   `smill build typst`; add `TYP` codes to `docs/cli.md` and `docs/typst.md`.
5. Add `.typst-version` and a version check mirroring Hugo's; CI may install a
   checksum-verified Typst for integration only after approval; no artifact upload.
6. With separate generation approval: compile PDFs locally, inspect, run the
   accessibility checker, record in a verification document.

Implement in its own `codex/005-*` branch/PR. PDF generation, distribution builds
and publication each remain separately gated. Do not create placeholder CLI
functionality before the renderer exists.
