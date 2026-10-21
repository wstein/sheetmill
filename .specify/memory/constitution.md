# SheetMill Constitution

Version: 1.2.0 · Ratified: 2026-10-13 · Last amended: 2026-10-21

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
7. **Human publication control.** Scope extensions require human authorization,
   recorded in `AGENTS.md` and the affected feature. The approved scope includes
   core/Reference/Hugo, navi text export and read-only Studio. The named Pages
   manual/demo workflow has its own recorded approval. The tag-driven release
   workflow (`.github/workflows/release.yml`, Spec 019) publishes only GitHub
   Releases of `wstein/sheetmill`, only on an annotated version tag `vX.Y.Z` pushed
   by the repository owner after the verify gates pass. Each tag and release needs
   the owner's own explicit approval for that tag, given by pushing it: there is no
   standing approval, a `workflow_dispatch` dry run never publishes, and the approval
   is not recorded per version in `AGENTS.md`. Broader publication stays
   gated. Typst/PDF, browser editing and playgrounds remain unapproved. Merges,
   tags, releases, distribution builds and publication require their separate
   approvals; a spec or automated check grants none. No navi, source command or
   dynamic suggestion executes.
8. **Review evidence.** Report actual verification and gaps. Independent review is
   expected before merge; passing author tests do not satisfy that expectation.

Amendments require a human-approved change to this file plus an impact check on
format docs, feature specs, code and tests. No automation may bypass permissions.

## Amendment impact (2026-10-14)

The user requested spec-status reconciliation and 006 T009 implementation. This
minor revision reconciles authorization wording with later recorded approvals;
determinism, canonical source, inert inputs and human publication control remain.
Impact checked against AGENTS.md, the feature index, 006/007/012 specs/tasks,
Studio roadmap and format docs. This amendment changes no runtime behavior or
verification result, and does not treat any unfinished study or gate as complete.

## Amendment impact (2026-10-21)

The owner (Werner Stein) approved preparing this change on 2026-10-21; his merge of
the pull request is the human-approved amendment. Principle 7 now names the
tag-driven release workflow of Spec 019 beside the Pages workflow (open decision 1
of that spec) and states that the approval of a tag is the owner's own tag push,
not a line per version in `AGENTS.md`. The "current T009 increment" sentence was
stale (T009 merged in #30) and is removed. Impact checked against AGENTS.md, the
019 spec, the feature index, `docs/releasing.md` and the other docs that mention
releases. This amendment changes no runtime behavior and grants no tag, release or
workflow run.
