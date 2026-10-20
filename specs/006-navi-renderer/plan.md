# Plan — 006

Implemented in `internal/render/navi` and `internal/cli/navi.go`; the CLI shares
file selection and parsing with `smill build hugo` (`loadSelection`). The IR gained
`id_explicit`, an additive field omitted when false. See [spec](spec.md), the
[format page](../../docs/format/navi-profile.md) and [source evidence](research.md).

1. Contract and decisions: spec, format page, immutable upstream evidence.
2. Tests: renderer table tests, golden bytes, grammar-model and deduplication
   checks, fuzz target, no-execution lint, CLI boundary tests.
3. Independent review of the final revision, then human merge. Publishing generated
   cheats, releases and distribution builds remain separately gated.
4. T009: emit deterministic schema-1 metadata companions beside each exported sheet;
   preserve all frontmatter and list actual exported cards.
5. Later, each needing its own decision: stronger syntax
   validation and a suitable Fish parser (Bash/Zsh quote diagnostics use pinned
   `mvdan.cc/sh/v3/syntax` v3.14.1), shared-context inheritance, nushell raw strings (choice lists and suggestion
   commands moved to Spec 020).

No navi, finder, shell or source snippet runs in renderer tests or CI.

## 015 integration

[015](../015-command-cards/plan.md) owns core command/parameter semantics and
migration. 006 consumes the shared contract under the 2026-10-14 implementation
handoff, retaining profile/payload/placeholder and NAV008 identity rules.
Complete byte parity and fixed pinned hashes are regression requirements.
