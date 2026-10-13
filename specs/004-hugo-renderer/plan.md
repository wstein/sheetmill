# Plan — 004

Use `internal/render/hugo` for pure IR-to-HTML/project generation, embedded original
Hugo templates and static CSS/JS. The CLI owns explicit file discovery, parsing,
route selection and exclusive output creation. It never launches Hugo or downloads
anything; the documented second command invokes pinned Hugo locally. The complete
workflow produces actual HTML output, rather than a placeholder build command.

Export one generated frontmatter-only page per sheet and trusted, pre-escaped HTML
in Hugo data JSON. Hugo templates never receive source Markdown. `safe.HTML` is
restricted to the renderer-owned HTML field. Other metadata is escaped normally.
Generate stable section/card anchors with a content namespace and rewrite recognized
local fragments. Search JSON contains plain text plus generated local destinations;
JS uses textContent/createElement, never source innerHTML.

Use root-relative local URLs (v0.1 sites served at the origin root), a fixed build
clock, disabled RSS/taxonomies/sitemap and denied Hugo subprocess/network security
settings. No user config/theme extensions are imported. Unsupported nodes and unsafe
links fail; inert HTML/image transformations warn rather than silently losing content.
Retain metadata in an exported canonical data file and source/license disclosures.

Brand source is embedded from the approved asset package; exports use pinned font
hashes and local tools. Serve an OFL-licensed local IBM Plex Sans webfont. Code uses
an explicit system monospace stack; no PDF font selection is implemented.

Verification: failing behavior tests, deterministic semantic goldens, optional actual
Hugo integration gated by explicit SHEETMILL_HUGO (skip if absent), repeated output
hashes, browser desktop/mobile/search checks, race/vet/fuzz and Sol exact-SHA review.
