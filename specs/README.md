# Feature sequence and status

Reconciled 2026-10-14 against main `0d42eb2` and the recorded review/verification
links below. Status is evidence of delivered work, not permission for a new action.
Current authorization is recorded in [AGENTS.md](../AGENTS.md); a feature draft
never grants execution, merge or publication permission.

| Feature | Delivered status | Remaining scope / boundary |
| --- | --- | --- |
| [001](001-core-foundation/spec.md) | Foundation implemented and merged ([#1](https://github.com/wstein/sheetmill/pull/1)) | Historical review checkboxes require their own linked evidence |
| [002](002-gfm-parser/spec.md) | GFM+ parser and semantic IR implemented ([#2](https://github.com/wstein/sheetmill/pull/2)) | Historical review checkboxes require their own linked evidence |
| [003](003-reference-import/spec.md) | Controlled Reference conversion implemented ([#3](https://github.com/wstein/sheetmill/pull/3)) | Provenance does not grant redistribution rights |
| [004](004-hugo-renderer/spec.md) | Hugo renderer and local generation implemented | Hosting authorization is scoped separately by 012 |
| [005](005-typst-renderer/spec.md) | Specification only | Version/fonts/accessibility and layout decisions pending; no implementation/PDF generation |
| [006](006-navi-renderer/spec.md) | Exporter and review fixes merged ([#27](https://github.com/wstein/sheetmill/pull/27), [#28](https://github.com/wstein/sheetmill/pull/28)) | T009 sidecars merged in [#30](https://github.com/wstein/sheetmill/pull/30); stronger syntax/provider work remains pending; never execute navi |
| [007](007-live-preview/spec.md) | Read-only serve, reader corpus and automated performance checks implemented | T049 five-author usability study remains pending; no browser editing |
| 008–010 | Reserved; no feature directories | Visual editing, disposable and online playgrounds remain proposals ([roadmap](../docs/studio-roadmap.md)) |
| [011](011-heading-attribute-ir/spec.md) | Specification only | Resolve heading-boundary and IR schema decisions before separate implementation authorization |
| [012](012-public-hosting/spec.md) | Base path, approved manual/demo hosting and canonical sitemap implemented | General redistribution gate and header/CSP follow-up remain pending; no deployment in this increment |
| [013](013-tldr-import/spec.md) | Specification only | Pin source evidence and settle conversion decisions before implementation authorization |
| [014](014-library-navigation/spec.md) | Implemented and locally verified ([#33](https://github.com/wstein/sheetmill/pull/33)); Sol review clear; CI readiness tracked on PR | Per-sheet group/order first slice; manifest, titles, IDs and deployment deferred |
| [015](015-command-cards/spec.md) | Shared command cards and migration implemented; local verification and Sol code re-review pass | Core `.command` / `cheatsheet.cards` implemented under the 015 handoff; 006 consumes it with unchanged navi bytes/hashes; no navi execution or deployment |
| [016](016-reference-cards/spec.md) | Draft PR #35; review round 1 fixes pending re-review | Opt-in `--theme reference-cards` layout; scroll dots, highlighting, footer and deployment excluded |

The public source contract is [docs/format](../docs/format/gfm-plus.md). Each feature
has spec, plan and tasks; use `SPECIFY_FEATURE_DIRECTORY=specs/NNN-name` with Spec
Kit scripts, and isolated `codex/` branches with focused Conventional Commits.
No unchecked review task is completed merely because a feature was merged.

006's combined review was recorded at candidate `caf1fc6` in
[#25](https://github.com/wstein/sheetmill/pull/25#issuecomment-6160173135); that PR
closed without merging. The implementation and fixes reached main through #27/#28.
These integration SHAs differ from the reviewed candidate; the review record keeps
the exact revision instead of silently relabeling it as another SHA.

007's [plan](007-live-preview/plan.md), [tasks](007-live-preview/tasks.md),
[Studio verification](../docs/studio-verification.md) and
[corpus evidence](../docs/e2e-corpus.md) distinguish implemented increments from the
pending human study. The active Spec Kit pointer is `.specify/feature.json`;
it does not override these feature boundaries or authorize all unchecked tasks.
