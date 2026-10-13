# Plan — 001

Use Go 1.27.2 after checking sibling toolchain policy and current Go support.
Initialize main with only the license, then work on `codex/001-core-foundation`.
Use `cmd/smill` and a testable `internal/cli.Run` entrypoint. Keep only two justified
parser dependencies; pin exact module versions and Actions commits.

Initialize Spec Kit 1.1.3 in a scratch directory, inspect and copy only generated
`.specify`/`.agents` project files, retaining upstream MIT license. Write feature
spec/plan/tasks explicitly; avoid `--force` over the populated repository.

Public source-format documentation lives in `docs/format`; specs track implementation
acceptance. Inspect sibling brand guides without copying their artwork. Document
contrast measurements and ratings; palette selection remains human-owned.

Verification: format, test, vet, module verification, CI, GitHub protection readback.
