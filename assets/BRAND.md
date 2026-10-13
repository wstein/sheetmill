# SheetMill brand specification

**Status:** Direction and palette selected — Editorial Stack, revision 2 (2026-10-13). Logo approved by the user on 2026-10-13; final export suite authorized.
**Product:** **SheetMill** · **CLI:** `smill` · **Tagline:** **One source. Every reference.**

This document records the selected brand direction and palette. The user requested the same three-sheet stack concept with updated colors and correction of the review findings on 2026-10-13. The user approved the presented logo geometry on 2026-10-13 and authorized final brand exports. The approved normal-size geometry is locked. Small-size optical adaptation and export derivatives retain the same three-sheet identity. No unseen concept-board illustration or sibling artwork was copied. Raster exports, banners, social cards and local Hugo generation are now explicitly authorized. Deployment, PDF and navi output remain separately gated.

## 1. Brand relationship and character

SheetMill is an independent, loosely coupled fourth product alongside BundleDock, WorkHarbor and CrewBook. It shares the family's disciplined typography, recognizable icon silhouette, reproducible vector artwork, and concise copy, **not** WorkHarbor's navy/teal palette or anchor icon.

**Positioning:** A Git-native Go compiler for authoring technical reference sheets in GFM with YAML frontmatter with planned outputs for web, multi-page PDF (Typst), and terminal (navi). The current implementation covers checking, inspection and controlled Reference import.

**Personality:** editorial, crisp, approachable, technical, calm. Avoid corporate gradients, 3D effects, physical-paper realism, glossy shadows, and overly playful cartoon shapes.

## 2. Names and voice

| Context | Spelling |
| --- | --- |
| Display name, title, wordmark | `SheetMill` (capital S and M, no space) |
| Repository, Go module suffix, package paths | `sheetmill` |
| CLI / executable | `smill` |
| Default tagline | `One source. Every reference.` |

Recommended product descriptor: “Git-native cheatsheet compiler.” Use “reference” broadly and “cheatsheet” for the main user-facing format. Do not imply all commands are executable or that every imported document is safe to run.

## 3. Palette

### Brand colors

| Token | Hex | Purpose |
| --- | --- | --- |
| `coral-500` | `#EA877F` | **Dominant primary**, primary symbol fill, `Mill` wordmark accent, major highlights |
| `coral-700` | `#B34E49` | Accessible accent text and interactive foreground on light backgrounds; validate contrast |
| `coral-100` | `#FDF0EE` | Pale wash, backgrounds and badges |
| `periwinkle-500` | `#7C8CFF` | Rear sheet and decorative non-text accents; not essential controls on white |
| `periwinkle-700` | `#4353B8` | Accessible secondary links/text on light backgrounds; validate contrast |
| `periwinkle-100` | `#E9ECFF` | Secondary pale wash |
| `ink-900` | `#2C3140` | Primary type; use `dark-surface` for the icon tile |
| `slate-600` | `#5F6678` | Secondary text on light backgrounds; validate contrast |
| `mist-50` | `#F7F8FC` | Light page background |
| `white` | `#FFFFFF` | Card surfaces and reverse type |
| `border-200` | `#D9DDF0` | Subtle separators |

The primary is **coral `#EA877F`**, not periwinkle. Use periwinkle sparingly for the third sheet, small accents, and secondary actions. In a colored symbol, coral should occupy at least 60% of the visibly filled area (excluding transparent negative space and any backing tile). No gradients in primary marks or icons.

**Contrast policy:** Coral `#EA877F` and periwinkle `#7C8CFF` are below 3:1 on white and must not be used for ordinary text at any size or essential graphical controls on white. Use ink for body text; use the darker color tokens for links and controls after testing WCAG contrast for the exact foreground/background pair. Require WCAG AA (4.5:1 for normal text, 3:1 for large text and essential graphical controls). Never communicate important meaning by color alone. Coral-filled controls should generally use **ink** lettering unless contrast tests support another pairing.

### Verified contrast pairs

Calculated from sRGB relative luminance; threshold decisions use unrounded values.

| Foreground / background | Ratio | Approved role |
| --- | --- | --- |
| `coral-500` / white | 2.54:1 | Logo/decorative artwork only |
| `periwinkle-500` / white | 2.98:1 | Logo/decorative artwork only |
| `coral-700` / white | 5.12:1 | Normal accent text |
| `coral-700` / mist | 4.82:1 | Normal accent text |
| `periwinkle-700` / white | 6.63:1 | Normal secondary text |
| `slate-600` / white | 5.74:1 | Normal secondary text |
| `ink-900` / coral-500 | 5.10:1 | Coral-filled button labels |
| `coral-500` / dark-surface | 4.38:1 | Logo/decorative artwork or large text; not normal text |
| `periwinkle-500` / dark-surface | 3.74:1 | Decorative artwork or large text; not normal text |

A genuine brand wordmark is exempt from the text-contrast criterion. This exception
must not be extended to taglines, ordinary headings, navigation or control labels.
Use `logo-accessible.svg` for the darker coral wordmark variant on light surfaces;
use ordinary ink/white text for small UI names. The wordmark exception does not
waive other applicable link/control accessibility requirements.
[W3C text contrast](https://www.w3.org/WAI/WCAG22/Understanding/contrast-minimum.html),
[W3C non-text contrast](https://www.w3.org/WAI/WCAG22/Understanding/non-text-contrast.html).

### Dark surfaces

| Token | Hex | Use |
| --- | --- | --- |
| `dark-surface` | `#343B50` | Canonical dark app-icon tile / dark UI surface — deliberately lighter than WorkHarbor navy |
| `dark-card` | `#40485E` | Elevated panels |
| `dark-text` | `#F7F8FC` | Main reverse text |
| `dark-muted` | `#D0D5E5` | Secondary reverse text |

Dark mode is a presentation adaptation, **not** a second brand. Preserve coral and periwinkle with sufficient distinction against dark surfaces. Avoid reusing WorkHarbor's `#0B2545` navy.

## 4. Primary icon — Editorial Stack

**Concept:** three staggered paper/reference cards, flat and ordered. The foremost page is coral, the middle page pale coral, and the rearmost page periwinkle. This is **not** a monogram-S, folded-corner emblem, book, anchor, or gear.

### Canonical geometry

Build the master SVG on a **64 × 64 viewBox** (`viewBox="0 0 64 64"`). This grid defines proportions, not the final image pixel size.

- Candidate source geometry is recorded in `tokens.json`; rotation is exactly 0 degrees.
- Canvas padding / safe-area target: at least 6 units on all sides; icon bounding box within `x=7..57`, `y=7..57`.
- Each sheet is an upright **rounded quadrilateral / near-rectangle** with consistent corner radius approximately 4 units; no stroke and no shadow.
- Front coral sheet: dominant at **upper-left**, visible as the largest solid page. Bounding box approximately `(9, 9)` to `(43, 47)`.
- Middle sheet fill is exactly `coral-100` (`#FDF0EE`). Offset about **6 units right and 5 units down**, with a clearly visible lower/right edge.
- Rear periwinkle sheet: offset another **6 units right and 5 units down**, visible primarily at the lower-right edge.
- Layer order (back to front): periwinkle, pale coral, coral. Expose the 6-unit right and 5-unit lower offsets; do not add invented gaps, borders or shadows.
- Angle: a *slight* unified lean (about 8–12 degrees) is permitted; do not use perspective distortion. Corners and edges should share a consistent angle and curvature.
- No text, border, gradient, embossing, or decorative lines inside the symbol.

**Geometry status:** the approved normal-size SVG masters use exact rectangles at (21,19), (15,14) and (9,9), all 34 × 38 with radius 4. Their order is rear, middle, front. The optional dark tile is (0,0), 64 × 64, radius 14, fill `dark-surface`. The approved normal-size geometry must not be changed silently. The small-size master uses 32 × 36 sheets at (21,20), (15,14), (9,8), radius 4; larger exposed edges preserve legibility.

### Small-size adaptation

- At **32 px and above**, use all three sheets.
- At **16–24 px**, use a separately optically adjusted SVG master: enlarge gaps / exposed edges; simplify tiny slivers rather than shrinking blindly.
- At **16 px**, coral must remain unmistakably dominant and the stack must not collapse into a single blob.
- Do not use different symbols across favicon, banner, README or CLI help.

## 5. Logo forms — two defined families, approved logo and derived export assets

**A. Icon-only** for favicons, CLI/package icons, avatars and square social placements: Editorial Stack symbol, either on a light surface with safe padding or centered in a rounded dark tile. The dark tile is *not* required for every light-context icon.

**B. Horizontal lockup** for README header, documentation navbar, social preview and presentations: icon to the left and the one-line **SheetMill** wordmark. Optional tagline is a separate line below the wordmark, not squeezed alongside the icon.

For consistency, treat a symbol-in-tile variant as a specific asset, never arbitrarily wrap a lockup symbol in a tile in CSS.

### Wordmark

- Font family: **Bricolage Grotesque**, outlined from the pinned source in `README.md`, with `wght=700`, `wdth=100`, `opsz=72`. SVGs contain no live text or external font dependencies; no font binaries are committed.
- “**Sheet**” in `ink-900` (`#2C3140`) on light surfaces and white in dark contexts.
- “**Mill**” in `coral-500` (`#EA877F`) on both surfaces **when used as large display wordmark**. This is a brand-wordmark exception, not permission for low-contrast large body text. The defined light accessible variant uses `coral-700`; small UI text uses ink/white.
- Letterforms are typeset as a **single joined word** with one baseline, one optical weight, and no gap at the capitalization boundary.
- Never stylize the wordmark as `SHEETMILL`, `Sheet Mill`, `Sheetmill`, `sheetMill` or `smill` (except the CLI itself).

### Proportions (relative to wordmark cap height C)

| Item | Rule |
| --- | --- |
| Icon visual height | `1.9 × C` in primary horizontal lockup |
| Icon-to-wordmark gap | `0.55 × C` |
| Wordmark line | Bricolage Grotesque Bold, single baseline |
| Optional tagline | IBM Plex Sans regular/medium, about `0.43 × C` cap height |
| Tagline placement | Begins at wordmark's left edge, below baseline with clear separation |
| Minimum clear space | `0.6 × C` on all sides of the full lockup |

Any optical adjustments to these ratios must be documented when committing final master assets, so exported variants remain consistent.

## 6. Typography

| Role | Family | Usage |
| --- | --- | --- |
| Brand wordmark and display headings | Bricolage Grotesque Bold | Logo, banners, large headings |
| UI body, captions, documentation | IBM Plex Sans | Explanations, navigation, tagline |
| Code, commands, key bindings | IBM Plex Mono | CLI snippets, GFM code fences, keyboard hints |

Use consistent weights and spacing. Future PDF output must select an explicit font profile that pins every family, file hash, weight and variation axis. A missing or mismatched font is an error; never substitute according to host availability. An alternate profile is a deliberate configuration change with layout regression checks. IBM Plex Sans/Mono profiles must be pinned before renderer implementation; no PDF is generated by this branding task. Keep headings short and comfortable in multi-column references.

## 7. Asset inventory and export rules

Standalone SVGs must be self-contained, carry an accessible title/description, and use outlined lettering. No external images, fonts, scripts or linked resources are permitted.

All raster assets must be derived from checked-in SVG source masters. Do not individually redraw exported PNGs. Do not rasterize text in core wordmarks if a portable outlined SVG is available.

| Asset | Intended dimensions / format | Notes |
| --- | --- | --- |
| `assets/logo.svg` | SVG; natural lockup ratio | Horizontal symbol + outlined SheetMill wordmark; candidate exists |
| `assets/logo-dark.svg` | SVG; same lockup ratio | White Sheet / coral Mill on a dark surface; candidate exists |
| `assets/logo-accessible.svg` | SVG; same lockup ratio | Ink Sheet / dark-coral Mill; candidate exists |
| `assets/preview.svg` | 1200 × 650 | Local comparison board; candidate exists |
| `assets/icon.svg` | 64 × 64 viewBox | Source icon; candidate exists |
| `assets/icon-small.svg` | 64 × 64 viewBox | Optically simplified small-size master |
| `assets/icon-dark.svg` | 64 × 64 viewBox | Defined dark tile; candidate exists |
| `assets/banner.svg` | 1000 × 220 viewBox | README banner; icon, wordmark, optional tagline |
| `assets/social-preview.svg` | 1280 × 640 viewBox | Social card with text and whitespace |
| `assets/favicon.svg` | SVG optimized for tiny sizes | Derived from `icon-small.svg` geometry |
| `assets/favicon-32.png` | 32 × 32 | Generated |
| `assets/favicon-16.png` | 16 × 16 | Generated |
| `assets/apple-touch-icon.png` | 180 × 180 | Generated, opaque canvas; platform masks corners |
| `assets/social-preview.png` | 1280 × 640 | Generated from SVG |
| `assets/banner.png` | 1000 × 220 | Generated from SVG |

Inventory entries other than the five logo/icon candidates and comparison board remain planned, not generated. The sizes above are **SheetMill proposals**, not inherited dimensions from WorkHarbor. Additional platform-specific variants can be added if justified, but must share the same master geometry.

### Prohibited transformations

No gradients, drop shadows, 3D extrusion, inconsistent sheet count, freehand rotations of the symbol, per-output re-coloring, unapproved outlines, or changing the wordmark font to approximate a screenshot. Background tints are permitted outside the logo mark.

## 8. Web cheatsheets and PDF language

The product identity must serve the content, not overwhelm it.

**Web:** mist page background, white cards, ink headings/body, coral for key section markers and primary action emphasis, periwinkle for modest supplementary accents. Reserve generous whitespace, visible code-copy controls and stable anchors. Keyboard shortcuts require monospace keycaps with textual labels; do not rely solely on iconography.

**Typst/PDF:** default white paper, ink text, thin pale-coral or periwinkle dividers. Multi-page sheets are normal. Avoid flood-filled coral cards and dark page backgrounds that waste ink. Make code and keys selectable text where possible. Preserve readable code line lengths and prevent clipped tables or orphaned headings.

**Terminal/navi:** use plain-text-compatible labels and command descriptions. Do not attempt to carry decorative brand styling into shell output; preserve accurate command content and syntax.

## 9. Copy and tone

Preferred headline: **SheetMill**.
Primary tagline: **One source. Every reference.**
Long descriptor: “Compile Git-native Markdown cheatsheets into web, print, and terminal references.”

Prefer concise, factual, technical copy. Avoid claims such as “pixel-perfect across every output” or “automatically executable” unless independently demonstrated. The browser, PDF and navi backends have different capabilities.

## 10. Accessibility and visual QA gates

Before locking or publishing visual assets (candidate source/preview work is allowed within the user’s request):

1. Review the current light/dark lockups and 64/32/16-px icon samples together. Extend the comparison board with banners/social cards when those assets are explicitly requested. Renderer examples are a separate later gate.
2. Check coral dominance, shape consistency, wordmark proportions, font licensing, and whether the three-sheet silhouette remains recognizable at 16 px.
3. Verify WCAG AA contrast for **actual** interactive/text combinations, and include text labels where color conveys meaning.
4. Inspect PNG derivatives against the SVG master and record their generation commands and tool versions.
5. Inspect the identity in monochrome/grayscale before locking it. Actual PDF print/layout checks belong to the separately authorized Typst renderer milestone.
6. Obtain explicit human approval of the comparison board **before** treating visual assets as locked. Do not interpret approval of this prose specification as approval to build or publish assets.

## 11. Engineering and governance

- Keep `assets/BRAND.md` authoritative for brand decisions after review. No asset-generation script or theme token may silently override it.
- The current brand direction is **Editorial Stack** only; **Monogram Stack-S** is a separate alternative, not an interchangeable official logo.
- Version tokens centrally in `tokens.json`. Source SVG generation uses pinned tools/font hashes, no clocks or automatic downloads, and must reproduce byte-for-byte.
- SheetMill's brand assets and Go compiler are independent of BundleDock, WorkHarbor and CrewBook; reuse *principles*, not their logos.
- Use focused Conventional Commits for proposed changes, for example `docs(brand): define Editorial Stack identity` and a separate later `feat(brand): add canonical SVG masters`.
- Prepare and review source without assuming permission to build documentation, create release artifacts, deploy, merge, or modify the other repositories. Respect repository-level `AGENTS.md` and human approvals.

## 12. Approval record

- **Direction decision:** selected by the user’s stack-logo/color-update instruction on 2026-10-13: Editorial Stack (three-sheet icon)
- **Primary:** coral `#EA877F`
- **Secondary:** periwinkle `#7C8CFF`
- **Typography:** Bricolage Grotesque / IBM Plex Sans / IBM Plex Mono
- **State:** direction/palette selected; user approved the presented logo on 2026-10-13; final asset exports and local Hugo generation authorized
- **Next:** verify the finalized comparison/export inventory; obtain separate approval for subsequent design changes or deployment
