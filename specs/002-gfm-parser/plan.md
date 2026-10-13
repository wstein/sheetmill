# Plan — 002

Use Goldmark v1.8.6 with GFM and go.yaml.in/yaml/v3 v3.0.5 with explicit node-tree
conversion. Standard library owns CLI, IO boundaries, diagnostics and serialization.
Packages: `frontmatter`, `validate`, `diagnostic`, `markdown`, `model`, `cli`.

Retain original Markdown body for syntax fidelity plus a semantic AST independent
of upstream Go types. Interpret only root section/card headings; keep heading inline
nodes while excluding attrs from structural titles. IDs use explicit anchors or
source-order slugs. Source positions reference the complete original bytes.

Do not normalize code or run snippets. Keep metadata schema separate from the IR
version. Strict unknown-heading-attribute behavior precedes any permissive importer.
Verify with table-driven failures, semantic assertions, snapshots and bounded fuzzing.
