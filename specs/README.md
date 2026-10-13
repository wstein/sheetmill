# Feature sequence

| Feature | Scope | Current authorization |
| --- | --- | --- |
| [001](001-core-foundation/spec.md) | Repository, Go, CLI foundation, policy and CI | Implement |
| [002](002-gfm-parser/spec.md) | Strict GFM+, semantic IR and checks | Implement |
| [003](003-reference-import/spec.md) | Controlled Reference conversion and provenance | Implement |
| [004](004-hugo-renderer/spec.md) | Hugo reference website | Specs only |
| [005](005-typst-renderer/spec.md) | Typst multi-page PDFs | Specs only |
| [006](006-navi-renderer/spec.md) | Safe annotated command export | Specs only |

Public source compatibility is defined in `docs/format`, not redefined per feature.
Each feature has spec, plan and tasks. Use `SPECIFY_FEATURE=001-core-foundation`
(or the selected feature name) with Spec Kit scripts when working on the combined
`codex/001-core-foundation` branch. Renderer task lists record future work, not
approval to execute it. Multiple focused conventional commits keep current work
reviewable. Independent review remains pending until separately recorded.
