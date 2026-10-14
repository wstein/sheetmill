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
   `mvdan.cc/sh/v3/syntax` v3.14.1), choice lists without silent executable providers,
   shared-context inheritance, nushell raw strings.

No navi, finder, shell or source snippet runs in renderer tests or CI.

## Pending 015 integration

[015](../015-command-cards/plan.md) defines core command/parameter semantics and
migration. 006 will consume that shared contract after separate implementation
authorization, retaining its profile/payload/placeholder and NAV008 identity rules.
No core/parser/schema changes are authorized by this documentation revision.
Complete byte parity and fixed pinned-hash acceptance precede alias integration.
