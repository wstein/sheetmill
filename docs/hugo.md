---
title: "Hugo website export"
description: "Exporting, building and hosting a reference library with Hugo."
category: guide
---
`smill` exports source for **Hugo 0.167.0**. It does not install or launch Hugo.
Use an explicitly installed tool and a **new** output directory:

```sh
go run ./cmd/smill build hugo ./examples/sheets --out ./.local/my-library --title 'My reference shelf'
hugo version
hugo --source ./.local/my-library --destination public --clock 2000-01-01T00:00:00Z --noBuildLock
python3 -m http.server 1313 --bind 127.0.0.1 --directory ./.local/my-library/public
```

Create `.local` first (`mkdir -p .local`). Hugo resolves the relative destination
under its source directory. Browse <http://127.0.0.1:1313/>. The local server binds
only to loopback. This is local generation, with no deployment/release/upload.
Opening `index.html` as `file:` is insufficient for local search; serve the root.
By default the site is served at the origin root. To serve it under a path, pass
`--base-path /sheetmill/` (leading and trailing `/`, unreserved URL characters
only; invalid values fail with `REN003`). Every generated link, asset, search URL
and permalink then carries the prefix; the published file layout stays the same,
so the host maps the prefix to the `public` directory.

All inputs are explicitly selected, regular `.md` files or one-level directories.
Validation completes before creating output. Existing destinations (including
symlinks) are refused; no overwrite flag exists. IO failure after creation can
leave a partial new project, reported as `IO003`. No existing project is removed.
Routes use lowercase ASCII filename slugs; all-non-ASCII names use a deterministic
SHA-256 prefix. Collisions, duplicate basenames and reserved `assets`/`brand`
routes fail with `REN003` rather than overwriting pages. Duplicate selected paths
are deduplicated. The output parent must already exist and have no symlink components.

The renderer consumes schema-1 IR without reparsing the original source. It writes
pre-escaped HTML to `data/sheets`, canonical complete metadata/source/IR to
`data/ir`, frontmatter-only pages, original templates and local static assets.
Hugo's `safe.HTML` is limited to the renderer-owned body. Source text is never
passed through Hugo Markdown, shortcodes or template evaluation. Inline HTML is
shown escaped with `REN101`; images become links with `REN102`. No source snippets,
plugins or linked resources are executed/fetched. Allowed links are relative,
HTTP, HTTPS and mailto; unsafe URLs fail with source spans (`REN002`). Selected
basename links such as `editor.md#select` become local page/card URLs. Recognized
structural IDs are namespaced with `ref-`, distinct from layout IDs `app-`.
Internal links that name no selected sheet, or a fragment that matches no section
or card ID of the target sheet, are kept as written and reported with `REN103`.

The library includes responsive ordered cards, GFM tables/tasks/code, local search,
code copy, an explicit dark-mode control, visible keyboard focus and print CSS.
Without JavaScript, navigation/content remain available; search/copy/theme controls
are progressive enhancements. Copy uses a compact icon beside the first code
line, without a reserved header row. Keyboard access, tooltip labels and announced
success/failure feedback are available; copying preserves the exact code text and
uses the browser clipboard only after activation.
The theme preference is stored locally; search runs entirely in the browser.
No CDN, analytics, remote theme or resource fetches are present. Source/license
metadata is visible on each sheet when supplied; presence of a license identifier
does not itself grant redistribution rights. Local IBM Plex Sans has its OFL
notice under `/brand/licenses`; the web UI deliberately uses a system monospace
fallback for code, independent of any future PDF font profile.

Verify the actual installed tool and build output explicitly:

```sh
SHEETMILL_HUGO=/absolute/path/to/hugo go test ./internal/render/hugo -run TestHugoIntegration -count=1
```

Without `SHEETMILL_HUGO`, Go tests do not invoke Hugo. The integration test compares
all built file bytes across two independent source/output directories with a fixed
clock. Unit tests cover semantic fidelity, escaping, unsafe URLs, unknown IR,
deterministic input ordering and output boundaries. CI installs only the pinned,
checksum-verified development tool; the check workflow uploads no site artifacts.

The [007 MVP harness](../tests/e2e/README.md) selects only
`examples/sheets/smill.md`, exports two fresh projects through the actual CLI,
compares every built public file byte, and automates desktop/mobile/no-JavaScript
reader journeys. This complements the renderer integration test without adding a
product preview server. See [actual E2E verification](e2e-verification.md).
The explicit `SHEETMILL_E2E_CORPUS=all` selection adds independently authored
BundleDock and WorkHarbor references to that same harness, including cross-card
links. See [verified corpus identities](e2e-corpus.md).

## GitHub Pages

Authorized on 2026-10-13 and extended on 2026-10-14 (AGENTS.md, specs/012).
`.github/workflows/pages.yml` runs on pushes to `main` and manual dispatch. It
repeats the core checks, then builds with the pinned Hugo:

- the manual from the pages listed in `docs/manual-pages.txt`, at
  <https://wstein.github.io/sheetmill/>, with `--source-base-url` and
  `--edit-base-url` pointing at `docs/` on GitHub;
- the independently authored `examples/sheets` demo at
  <https://wstein.github.io/sheetmill/demo/>;
- `llms.txt` and `agent-guide.md` as plain files at the site root.

It deploys through the `github-pages` environment, which accepts `main` only.
A test builds the manual pages and fails on any warning. Other sheets need the
012 redistribution gate first.

## Search engines and sharing

Every page carries a meta description (the sheet's `description`, else the site
`--description`) and Open Graph title, type and description. With
`--site-url https://host/<base-path>/`, whose path must equal `--base-path`, pages
also get a canonical link, `og:url` and `og:image` (the brand social preview), and
the export adds `sitemap.xml` with the home page and every sheet; it has no dates,
so output stays deterministic. No `robots.txt` is generated: crawlers only read it
at a host root, which a project site under a path does not control.

## Sheets in several directories

Each selected file keeps its path relative to the deepest directory that contains
the whole selection, so `format/gfm-plus.md#ids` or `../cli.md` resolve between
selected directories and edit links keep subdirectories. With
`--source-base-url https://github.com/<owner>/<repo>/blob/<ref>/<dir>/` (the
selection root), relative links that name no selected sheet point at the
repository file instead of raising `REN103`; links leaving the repository still do.


## Read-only live preview

`smill serve [file-or-directory] --hugo /absolute/path/to/hugo` reuses the same
Go parser, semantic IR, renderer and pinned Hugo in owned temporary projects.
Save files externally; the local authenticated browser shows revision state,
source diagnostics and last-valid output. Trusted Studio assets/publication
identity are injected into HTTP HTML responses only; exported/generated file
bytes are unchanged. See [CLI usage](cli.md#local-saved-file-preview),
[security boundary](studio-security.md) and [actual verification](studio-verification.md).
No deployment is part of starting a Studio session.

## GitHub page editing

The demo's **Edit on GitHub** link opens the authored SheetMill Markdown,
for example [bdock.md](https://github.com/wstein/sheetmill/edit/main/examples/sheets/bdock.md).
It is separate from `source.repository/path/revision`, which records the upstream
implementation used to write the reference. Provenance remains preserved verbatim.

For your own library, opt in explicitly when exporting:

```sh
smill build hugo ./sheets --out ./new-project --edit-base-url https://github.com/OWNER/REPOSITORY/edit/main/sheets/
```

The base must be an HTTPS `github.com/owner/repository/edit/ref/` URL with an
optional directory, without credentials, query, fragment or traversal. Each
selected source basename is URL-escaped and appended. Select the repository,
branch and directory containing the actual authored sheets; input files from
unrelated directories should not share this option. Invalid configuration is
reported as `REN003` before output is created. No destination is inferred from
source metadata, and generic exports omit the link by default. Links do not fetch
GitHub resources automatically or execute any source content.

The existing main-only Pages workflow supplies the SheetMill edit base. This fix
only changes its export configuration; no deployment was run while preparing it.
Local actual-Hugo verification at `/sheetmill/` passed desktop/mobile and
JavaScript-disabled checks for the visible bdock edit destination, unchanged
BundleDock provenance and zero off-origin resource requests.

### Source disclosure presentation

Rendered pages show attribution, license and **Source details** together in one
compact disclosure summary, wrapping only when needed. Full source metadata and the rights note remain inside the closed disclosure.
Metadata renders as escaped key/value rows, with Attribution and License labels.
Nested values use readable YAML text, so unknown fields are preserved without a
JSON block. Nonprinting controls use explicit YAML escapes to avoid HTML
normalization; ordinary multiline values remain readable. Values remain display text; repository URLs are not fetched. The native disclosure works without
JavaScript. The generated **Edit on GitHub** action appears alongside the reference
navigation, with an accessible label identifying the cheatsheet edit destination.
Neither presentation changes cheatsheet Markdown or provenance values.

### Fingerprinted CSS and JavaScript

Each export names its bundled stylesheet and script with the full SHA-256 digest
of their exact bytes: `assets/site.<digest>.css` and `assets/app.<digest>.js`.
Hugo templates use those paths from generated configuration, prefixed by the
selected base path. Unversioned `site.css` and `app.js` aliases are not exported.
Changing either bundled asset produces a new URL for that asset; unchanged bytes
keep the same URL across sheets, export directories and site prefixes. Fonts
remain local and CSS-relative font URLs keep working.

This prevents newly loaded HTML from reusing an older cached asset under the same
URL. It does not force already open tabs or cached HTML to refresh, and it does
not configure the host's cache headers. No additional tools or dependencies are
needed; hashing happens in the Go renderer before Hugo runs.

Browser verification covers root and `/sheetmill/` mounts on desktop, mobile and
without JavaScript: URLs match the downloaded bytes, resources return successfully,
fonts load, and home/sheet pages reference the same assets. Existing real-sheet
journeys continue to verify search, appearance and code-copy behavior.

## Classic compact theme

`--theme classic-compact` selects a denser layout that reads like a reference card:
a 48px top bar with search, a compact index home, ruled section bands, bottom-ruled
cards in an auto-fill grid, a sticky jump row, an in-page card filter (`/` to focus,
Esc to clear; non-matching cards are dimmed and the count is announced) and a
Compact/Comfortable density toggle stored in the browser. Print uses A4 landscape
in three columns. Editorial Stack remains the default (`--theme editorial`).
The layout adds one fingerprinted stylesheet, uses system fonts (no new font files),
needs no inline styles, and works without JavaScript. Open decisions: A4 vs Letter,
the minimum column width (270px) and whether to ship IBM Plex Mono/600.

## Reference cards theme

`--theme reference-cards` keeps the classic top bar, search, in-page card filter and
density toggle and restyles the cards as soft rounded tiles. The card title becomes
a badge in the tile's top-right corner. Inside a card, a paragraph followed by a
fenced block reads as a tinted description row over a command row (other paragraphs stay plain), and `bash`,
`sh`, `zsh`, `shell` and `fish` fences show a `$` prompt drawn by CSS, so it is
neither selectable nor copied. Section headings carry a `#` marker and the home page
title is centered. The library is a collapsible **Browse library** panel above the
content instead of a sidebar; it keeps grouping, `navigation.order` ordering, the
current-page marker and the tag link, and works without JavaScript. Badge, rows
and surfaces use only palette tokens, so every `--palette` and light/dark mode
applies unchanged. The theme adds one fingerprinted `cards.css`, no fonts or
external requests; its print rules use three plain columns without tint, shadows or the Browse panel, but have not been checked in a real print preview. Editorial remains the
default, and `classic-compact` is unchanged. It is a styling variant only: it
does not change the source format and does not highlight command tokens.

## Tags and grouped navigation

Sheet tags link to generated result pages listing matching references and their
unique sheet count. `Browse tags` opens the complete tag index at `<base-path>tags/`.
These are ordinary static links: tag discovery and category disclosure controls
work without JavaScript. Tags match exact, case-sensitive source labels; duplicate
tags on one sheet do not duplicate its results. This increment supports one tag
per result page, not combinations of tags or card-level tagging.

A tag URL is `<base-path>tags/<ascii-stem>-<digest>/`: the stem is at most 32 ASCII
characters and the digest is the first 12 hexadecimal SHA-256 characters of the
original UTF-8 label. Punctuation, Unicode and differently cased labels therefore
have stable, distinct routes independent of source ordering or other tags. Route
collisions fail with `REN003`. The top-level sheet route `tags` is reserved.
Canonical URLs and the generated sitemap include the tag index/results when a
site origin is configured; tag pages exist without an origin as well.

Both layouts have a left library sidebar grouped by `navigation.group`, else
`category` (see [library placement](format/gfm-plus.md#library-placement); sheets
inside a group follow `navigation.order`); uncategorized
sheets appear under Other references. Native disclosure controls expand or collapse
groups using mouse, Enter or Space. The current sheet's group opens on navigation,
and its link has a visible marker and `aria-current="page"`. With JavaScript,
other disclosure preferences persist per site base path. At narrow widths the
library moves above the content and remains collapsible. Existing section jump
links remain in the content; the sidebar does not enumerate individual cards.

## Home page shelf cards

Each library card on the editorial home page shows facts the sheet already
declares, as plain text: the number of reference cards (`3 cards`, `1 card`;
omitted for a sheet without cards), `Tool: NAME` from `cheatsheet.tool` when set,
the description, up to three tag pills and the license. Pills are the first three
distinct `tags` in source order, so authors choose which tags lead; the sheet page
and tag pages still list every tag. Nothing is computed from other sheets, and all
text is HTML-escaped. The compact layouts (`classic-compact`, `reference-cards`)
add only the card count to their index rows. The cards need no JavaScript, request
nothing remote and do not alter print output.

## Color palettes

Palette, layout and light/dark appearance are independent:

```text
smill build hugo sheets --out local-site --theme classic-compact --palette dracula
```

| `--palette` | Reading surfaces and accents |
| --- | --- |
| `graphite-coral` (default) | Neutral dark charcoal or pale light surfaces, coral and periwinkle |
| `paper-ink` | Warm paper/ink surfaces and restrained coral |
| `midnight-periwinkle` | Deep navy or cool pale surfaces, periwinkle links |
| `monochrome` | Neutral black/white surfaces and grayscale accents |
| `teal-copper` | Teal surfaces/links with copper accents |
| `dracula` | Purple/pink accents on dark charcoal, plus a SheetMill light adaptation |

The Color palette selector remembers the browser choice independently of light/dark
mode and density. Without JavaScript, the configured palette still applies and
follows the system color-scheme preference. Unknown palette names fail with
`REN003` before writing. Print uses neutral readable colors.

Styles are implemented locally with no external requests or new dependencies.
Dracula's dark palette draws from its [published color values](https://draculatheme.com/contribute);
SheetMill's light adaptation, muted text and code surfaces are adjusted for
readability. Browser checks measure text, muted text, link and accent token contrast
against page, card and code surfaces at 4.5:1 or better in every palette/mode.
