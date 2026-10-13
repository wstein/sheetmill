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
| Indigo & Amber branding exploration | 9/10 | Distinct from sibling teal/navy; promising measured text contrast, human selection still required |

Recommended next implementation after review: design a narrow TLDR adapter or
select the first renderer based on the desired user output. Do not start either
under current authorization. Renderer features 004–006 remain specs only.
