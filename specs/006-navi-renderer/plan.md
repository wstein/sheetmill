# Plan — 006

Documentation refinement only. The proposed first slice is
`bash-free-input-v1`: explicit cards, one physical Bash command per card and
free-input parameters. See [spec](spec.md), [format](../../docs/format/navi-profile.md)
and [source evidence](research.md). No renderer, validator or CLI implementation
is authorized. No compatibility execution has occurred.

1. Review the profile, closed metadata contract and acceptance matrix. Preserve
   current generic frontmatter/IR behavior; schema-1 output changes need review.
2. After separate authorization, select/pin a non-executing Bash 5.2 grammar
   parser and expose explicit-ID provenance through a reviewed additive IR
   representation. Define sidecar schema 1 with complete metadata and selected
   card records. Do not reparse source Markdown inside the renderer.
3. Write failing selection, metadata, syntax and placeholder tests; independent
   golden bytes; pure pinned-navi parser/interpolation compatibility fixtures;
   deterministic output tests, fuzzing and transitive no-execution audit.
4. Implement `internal/navi` and the proposed `smill build navi` only after those
   prerequisites. Add validator/schema, source diagnostics and user documentation;
   no suggestion providers or alternate shells. Do not add placeholder commands.
5. Run all acceptance checks and record exactly what was tested. Obtain independent
   review of the final revision; generation and publication remain separately gated.

Use a dedicated `codex/006-*` branch/PR for implementation. No navi, finder, shell
or source snippet runs in renderer tests or CI. The one-line restriction is a
first-slice tradeoff, not a claim of full navi syntax or round-trip compatibility.
