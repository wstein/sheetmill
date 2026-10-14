# Feature 016: reference-cards Hugo layout

**Status**: Implemented in PR #35; independent review CLEAR at fdee8ff, P3 follow-ups applied afterwards.
**Authorization**: user, 2026-10-14 ("write a spec and implement reference-cards as a
reviewed PR"). Extends feature 004; no merge, deployment or source-format change.

## Goal

Offer a third, opt-in Hugo layout, `--theme reference-cards`, in which each card is a
soft rounded tile with its title as a corner badge and its command examples as stacked
description/command rows. `editorial` stays the default; `classic-compact` is unchanged.
The idea of a badge-and-rows tile came from reviewing a public cheatsheet site; no
branding, wording, logo, colors, fonts, scripts or external requests are copied.

## Requirements

- **FR-1** `--theme reference-cards` is accepted by `smill build hugo` and
  `hugo.Options.Theme`; unknown themes still fail with `REN003`. The rendered HTML
  of `editorial` and `classic-compact` stays byte-identical (the exported `baseof.html`
  template source differs; verified by building both revisions).
- **FR-2** The layout reuses the classic markup (top bar, search, in-page card filter,
  density toggle, section jump list) and adds a body class `reference-cards` and one
  fingerprinted `cards.css`, shipped only for this theme.
- **FR-3** Cards render as rounded tiles with a soft shadow. The card title (`h3`)
  is a badge in the tile's top-right corner, in flow so long titles wrap without
  covering content; wide tables scroll inside the tile.
- **FR-4** Inside a card, a paragraph immediately followed by a fenced block is a tinted
  description row and the fence is the white command row; other paragraphs stay plain. Styling is CSS only; the
  source Markdown and generated HTML are unchanged.
- **FR-5** Fences labelled `bash`, `sh`, `zsh`, `shell` or `fish` show a `$` prompt
  drawn by CSS `::before`, so it is never selectable text and never copied.
- **FR-6** The home page title and description are centered; section headings are
  accent-colored with a `#` marker.
- **FR-7** The library moves from a permanent sidebar to a collapsible "Browse
  library" disclosure above the content. It keeps the 014 grouping, ordering,
  current-page marker and tag link, and works without JavaScript.
- **FR-8** Palettes (#31) and light/dark modes apply unchanged. Badge text and
  surfaces use palette tokens only; a Go test checks that the description-row
  pairs (`--text` and `--link` on `--wash`, `--link` on `--bg`) meet 4.5:1 in every
  palette and mode. No new colors, fonts or external requests.
- **FR-9** Layout fits at 320px and stacks to one column on narrow screens. Print rules
  (never checked in a real print preview) use three columns, plain bordered tiles, no
  tint, shadows or Browse panel.

## Out of scope

Syntax-highlighted tokens, scroll-spy dots, promotional footers, sharing/social
buttons, source-format changes, new diagnostics, redistribution of third-party
styling and deployment.

## Decisions

1. Reference-cards is a variant of the classic markup, not a fourth template tree; the
   `classic` Hugo param therefore stays true for both and a second `cards` param
   selects the extra stylesheet and body class.
2. The Browse disclosure replaces the sidebar in this layout (user left this open; the
   recommended option was taken).
3. Badge colors reuse `--link` on `--bg`, which the Go contrast test covers.
4. Column spans are re-thresholded for this grid's 300/340px columns.
