# Local Hugo library

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
The v0.1 project uses root-relative URLs and must be served at the origin root;
subpath deployment is not supported by this milestone.

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
are progressive enhancements. Copy uses the browser clipboard only after a click.
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
checksum-verified development tool and uploads no generated site artifacts.

The [007 MVP harness](../tests/e2e/README.md) selects only
`examples/sheets/smill.md`, exports two fresh projects through the actual CLI,
compares every built public file byte, and automates desktop/mobile/no-JavaScript
reader journeys. This complements the renderer integration test without adding a
product preview server. See [actual E2E verification](e2e-verification.md).
