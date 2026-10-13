# Plan — 006

Specification only. Start implementation after separate authorization and after
002-gfm-parser is accepted. Build a renderer around canonical IR, maintaining portable
semantic metadata and renderer-owned layout. Pin the external tool before use.

1. Close open decisions in spec.md: shell profile, card-level metadata and typed
   parameter schema (documented in `docs/format` with a JSON schema change),
   placeholder escape and leading-dash rules.
2. Author independent fixtures and failing table-driven tests for every negative
   row, plus golden `.cheat` bytes and a pure Go parser of the output grammar.
3. Add the no-execution lint (no `os/exec`, `syscall`, `plugin` in the package).
4. Implement `internal/navi` and `smill build navi`; add `NAV` codes to
   `docs/cli.md` and `docs/navi.md`; add a fuzz target.
5. Verify, record review; generation of artifacts only with separate approval.

Implement in its own `codex/006-*` branch/PR. Actual navi use, distribution builds
and publication each remain separately gated. Do not create placeholder CLI
functionality before the renderer exists.
