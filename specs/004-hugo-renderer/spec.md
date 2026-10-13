# 004 — Hugo renderer

Created: 2026-10-13 · Status: implemented; merged via PR #3 (see tasks.md)
Depends on: accepted core/Reference import (001–003)

## User scenarios and acceptance

1. P1: An author exports explicitly selected sheets into a new Hugo project and
   runs the pinned local Hugo tool to produce a browsable static reference library.
   Navigation, section/card anchors, code, tables and source attribution survive.
2. P1: A reader browses responsive cards or searches references entirely from local
   assets. With JavaScript disabled, ordinary navigation and content still work.
3. P1: An author receives source-positioned diagnostics for unsafe URLs or unsupported
   IR nodes. Raw HTML is shown inertly and images retained as links, with warnings;
   no source snippet, shortcode, plugin or linked resource is evaluated/fetched.
4. P2: The approved Editorial Stack identity is finalized as reproducible light/dark,
   favicon, small-size, banner, social and print-independent monochrome assets.

## Requirements

- FR-001: Consume schema-1 IR directly; do not reparse original Markdown or pass
  untrusted source text through Hugo Markdown/shortcode evaluation.
- FR-002: `smill build hugo <file-or-directory>... --out <new-project>` writes a complete
  self-contained project. Inputs use existing explicit/nonrecursive boundaries.
- FR-003: Require a new output directory; refuse existing/symlink destinations and
  duplicate sheet routes. Validate all inputs before writing. No overwrite mode.
- FR-004: Render supported GFM semantics, readable code, task state and table alignment.
  Escape HTML, text, attrs and URLs; allow only relative/http/https/mailto links.
- FR-005: Keep deterministic source/project/search bytes independent of input order,
  clocks and absolute host paths. Pin Hugo 0.167.0; compare actual built file bytes.
- FR-006: Preserve all metadata in JSON and show source/license attribution on pages.
- FR-007: Provide responsive layout, progressive local search, visible keyboard focus,
  semantic navigation, code copy, dark mode and print-friendly CSS.
- FR-008: Use original embedded templates/styles/scripts and local licensed brand/font
  assets; no remote themes, CDN, tool downloads, snippet execution or deployment.
- FR-009: User explicitly authorizes brand exports, implementation and local site
  generation on 2026-10-13. Merge/release/deploy and Typst/navi remain separately gated.

## Success criteria

Semantic and negative tests pass. New-output boundaries preserve existing files.
Project and actual Hugo output hashes match across repeated generation. Local site
is visually checked at desktop/mobile, with search/navigation exercised. a separately assigned independent reviewer (recorded in docs/hugo-verification.md)
reviews the exact final revision; findings are fixed and re-reviewed. Human approval
remains required before any merge.
