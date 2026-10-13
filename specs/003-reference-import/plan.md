# Plan — 003

Implement `internal/importer/reference` with pure Convert(source, provenance).
The adapter returns canonical Markdown and a typed deterministic report. CLI owns
input/output and accepts explicit provenance arguments, `--dry-run`, `--report json`.

Use the core strict frontmatter decoder and GFM parser for validation. Preserve body
bytes exactly; unsupported canonical heading attrs are errors, not silently dropped.
Legacy `plugins` and visual metadata move into `source.original_metadata` with
warnings. Existing source provenance is refused rather than replaced. Serialize
metadata using JSON inside YAML boundaries: JSON mappings are valid YAML and Go's
sorted map keys provide stable output without an additional emitter contract.

Output uses exclusive creation under an explicitly opened parent root and rejects
symlink paths. No implicit date, network access or dynamic execution. Test actual
Reference Python from a pinned scratch file; record its hash and upstream revision,
while keeping it out of the public EUPL fixture collection.
