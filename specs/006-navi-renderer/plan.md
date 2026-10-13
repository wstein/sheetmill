# Plan — 006

Implemented in `internal/render/navi` and `internal/cli/navi.go`, sharing the
selection and parsing step with `smill build hugo` (`loadSelection`).

1. Contract and decisions: spec.md and docs/navi.md.
2. Tests: renderer tests, fuzz target and no-execution lint; CLI boundary tests.
3. Review, then merge only after independent approval. Publishing generated
   cheats, releases and distribution builds remain separately gated.

Later increments (T009) need their own decisions: choice lists must not introduce
executable providers silently, and shared-context inheritance needs a collision
policy beyond per-sheet `NAV007`.
