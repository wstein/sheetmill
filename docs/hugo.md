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
