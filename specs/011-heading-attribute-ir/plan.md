# Plan — 011

Move the suffix decision entirely into `internal/markdown`. Today `heading()` already
finds the group start in the joined contiguous tail; extend it to return the cut
point (source offset) and build the heading's inline nodes only from content before
it into a new title-node field, trimming trailing whitespace and truncating or
dropping the boundary `Text` node with exact spans; `heading` stays unchanged.
Escape handling reuses `escapedPunctuation`.

Decide schema handling first (open decision 1) and update `internal/model/model.go`,
`schema/cheatsheet.schema.json` and docs/format/gfm-plus.md before code. Remove `renderer.heading()` suffix logic from
internal/render/hugo and render the title nodes directly. Reference import re-parses
with `markdown.Parse`; confirm its fixtures stay byte-identical.

Verify with table-driven negative tests, a fuzz target, regenerated and inspected
golden JSON, and before/after Hugo byte comparison of existing fixtures. Document
any intended byte difference. Studio (proposed editor) depends on this separation
to preserve heading attributes when titles are edited.
