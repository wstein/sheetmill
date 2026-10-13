# Feature sequence

| Feature | Scope | Current authorization |
| --- | --- | --- |
| [001](001-core-foundation/spec.md) | Repository, Go, CLI foundation, policy and CI | Implement |
| [002](002-gfm-parser/spec.md) | Strict GFM+, semantic IR and checks | Implement |
| [003](003-reference-import/spec.md) | Controlled Reference conversion and provenance | Implement |
| [004](004-hugo-renderer/spec.md) | Hugo reference website | Implement + local generation; no deployment |
| [005](005-typst-renderer/spec.md) | Typst multi-page PDFs | Specs only |
| [006](006-navi-renderer/spec.md) | Safe annotated command export | Specs only |
| [007](007-live-preview/spec.md) | Studio local live preview; first slice GFM → Hugo → browser E2E | Specification only |
| 008–010 | Studio editor and playgrounds ([roadmap](../docs/studio-roadmap.md)) | Reserved; not created |
| [011](011-heading-attribute-ir/spec.md) | Parser-owned heading title nodes in the IR | Specs only |
| [012](012-public-hosting/spec.md) | Base path and publication gate for hosting | Base path + Pages deploy of manual and demo sheets; gate specs only |
| [013](013-tldr-import/spec.md) | Controlled TLDR page import | Specs only |

Public source compatibility is defined in `docs/format`, not redefined per feature.
Each feature has spec, plan and tasks. Use `SPECIFY_FEATURE_DIRECTORY=specs/001-core-foundation`
(or the selected feature directory) with Spec Kit scripts. Each feature uses its own
`codex/NNN-*` branch (for example `codex/004-hugo-renderer`). Feature 004 is explicitly authorized; features 005–006 still record future
work without implementation approval. Multiple focused conventional commits keep current work
reviewable. Independent review remains pending until separately recorded.

[Studio roadmap](../docs/studio-roadmap.md) records proposed 008 visual editing,
009 disposable playground and 010 static online playground. These feature directories
have not been created. The active feature is recorded in `.specify/feature.json`;
007 planning can proceed independently of the deferred 005/006 renderers.
