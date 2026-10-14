# Recommendations and ratings

Ratings reflect implementation value relative to scope and risk on 2026-10-13.
They are judgments, not measured adoption or guarantees.

| Recommendation | Rating | Reason / next action |
| --- | --- | --- |
| Canonical GFM+ plus isolated import adapters | 10/10 | Keeps renderers independent of external formats and makes losses testable |
| Strict core, explicit loss-aware Reference adapter | 9.5/10 | Predictable authoring plus practical migration; retain body/provenance and reject unsupported attrs |
| Spec Kit with numbered features and authoritative format docs | 9/10 | Good traceability; review feature scope instead of blindly running generated workflows |
| Small Conventional Commits and one independent approval | 9/10 | Reviewable implementation and clear merge control |
| TLDR adapter after the core/importer settles | 8.5/10 | Focused CLI examples; preserve placeholders and check content attribution separately |
| Navi importer after typed command metadata is designed | 7.5/10 | Valuable, but dynamic suggestions and shell semantics add substantial safety work |
| Broad automatic multi-ecosystem import now | 4/10 | Ambiguous recognition and conversion loss would widen the milestone too far |
| Editorial Stack with coral/periwinkle | 9/10 | User-selected direction; use dark accent tokens for ordinary text and review final SVG masters before locking geometry |
| KaTeX math rendering | Declined (2026-10-14) | Needs a vendored third-party library (~300 KB) against the asset budget and standard-library-first policy; importer archives `plugins: katex` |
| Offline mode (web manifest, service worker) | Declined (2026-10-14) | Low reader value for references, with cache-invalidation risk on path-mounted sites; revisit only with a concrete need |
| Tooltips (`data-tooltip`) | Declined (2026-10-14) | No Markdown syntax; raw HTML is shown inertly by design |
| Runnable snippets (third-party playgrounds) | Rejected | Conflicts with "nothing in a sheet is executed or fetched", adds CDN and supply-chain risk |
| Parser emits structural headings without the `{...}` attribute suffix | 7/10 | Follow-up: the Hugo renderer re-derives the suffix separately from `markdown.heading()` and the rules already differ (contiguous segments); changes IR/schema, so needs its own reviewed PR with golden updates |

Current sequence (2026-10-14): reconcile spec status, complete the authorized
006 T009 metadata sidecar (9.7/10), then refine 011's heading-boundary/IR schema
decisions before implementation (9.2/10). The 012 general publication gate remains
a separate follow-up (8.7/10); 007's five-author usability study can proceed
independently (8.5/10). Typst/PDF and playground implementation remain gated.
Navi export and read-only Studio are implemented; neither implies command
execution, browser editing or new deployment authorization.
