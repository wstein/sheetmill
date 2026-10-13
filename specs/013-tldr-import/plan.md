# Plan — 013

Add `internal/importer/tldr` with pure Convert(source, provenance, pathInfo) that
returns canonical Markdown and a typed deterministic report, mirroring 003. The
CLI adds `smill import tldr` beside `reference`, reusing the provenance flags,
exclusive output creation and `--report json` schema 1.

Parse line by line with the standard library: H1, `>` block, then example pairs.
Anything unmatched becomes a finding. Placeholders stay inside fences; prose
placeholders are code-spanned. Serialize frontmatter as sorted JSON in YAML
boundaries, as in 003. Revalidate with the core parser before writing.

Document `TLD` codes in `docs/imports.md`. Verify the page format and licenses
against a pinned upstream revision recorded outside the repository; fixtures are
authored independently. Add a fuzz target and golden conversion/report tests.
