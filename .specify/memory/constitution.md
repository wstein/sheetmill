# SheetMill Constitution

Version: 1.0.0 · Ratified: 2026-10-13 · Last amended: 2026-10-13

1. **Determinism.** Identical source bytes, configuration and tool versions MUST
   yield byte-identical canonical JSON/conversions. No implicit clocks or host paths.
2. **Canonical source.** Editable GFM+ and YAML frontmatter are authoritative.
   `docs/format/gfm-plus.md` defines the public contract; feature specs cannot
   redefine it. Schema changes require explicit compatibility decisions and tests.
3. **Loss-aware imports.** Import adapters MUST preserve supported content and
   provenance, report unsupported constructs and disabled behavior, and refuse
   to overwrite user files. No network fetches during parsing/import.
4. **Inert input.** Source content, plugin declarations, shell commands and dynamic
   variable suggestions MUST never execute during checks or conversion. Import
   compatibility does not grant permission to redistribute source material.
5. **Focused Go design.** Use standard library first, justified pinned dependencies,
   package boundaries, test-first development, golden fixtures and fuzzing.
6. **Independent product.** No runtime dependency on BundleDock, WorkHarbor or
   CrewBook. Original implementation and independently authored bundled fixtures.
7. **Human publication control.** Current authorization covers foundation, parser,
   Reference importer, approved brand finalization, Hugo renderer implementation,
   local site generation, checks and reviewable topic commits/PRs. Typst/navi
   remain specification-only. Merges, tags, releases, deployments, binary builds,
   PDF generation and artifact publication need separate approval. Hugo local
   site generation is explicitly authorized on 2026-10-13; deployment is not.
8. **Review evidence.** Report actual verification and gaps. Independent review is
   expected before merge; passing author tests do not satisfy that expectation.

Amendments require a human-approved change to this file plus an impact check on
format docs, feature specs, code and tests. No automation may bypass permissions.
