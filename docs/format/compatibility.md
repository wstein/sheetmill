# Compatibility boundary

SheetMill implements its own limited source contract inspired by Markdown reference
card conventions. This is **not** full compatibility with Fechin/reference, its
Hexo plugins, Tailwind classes or theme layout. GFM support comes from Goldmark;
compatibility means preserving document semantics in the supported subset.

Accepted headings, spans, IDs and inert `.wrap` are specified in
[gfm-plus.md](gfm-plus.md). The semantic classes `.note`, `.tip`,
`.warning` and reserved `.navi` belong to SheetMill's contract, without an upstream
Reference compatibility guarantee. Unknown heading attributes are errors in v0.1;
a permissive importer may be designed separately later.

`testdata/reference-compatible/` contains independently authored examples, not
copied upstream sheets. Fixtures cover ordinary content, tables/alignment, task
lists, links, nested lists/quotes/headings, longer code fences, literal shorter
fences, layout hints, unknown metadata and invalid conflicts. Each valid fixture
has a checked-in JSON golden snapshot. Snapshot tests also compare repeated parses.

No upstream GPL-3.0 implementation, CSS, plugin or authored sheet was copied.
Renderer layout and code-export behavior are outside this milestone.
