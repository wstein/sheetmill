# 001 — Core foundation

Created: 2026-10-13 · Status: implemented, review pending
Branch: `codex/001-core-foundation`

## User scenarios and acceptance

1. P1: A contributor can identify SheetMill, its license and supported toolchain.
   Given an empty public repository, foundation supplies a full EUPL-1.2 license,
   independent agent policy, pinned Go module, README and least-privilege CI.
2. P1: A user can run `smill version` with stdout reserved for requested data.
   The command emits `smill dev` plus LF; invalid commands return usage and exit 2.
3. P2: A maintainer can review separate conventional topic commits and numbered
   feature artifacts. No automated merge, release or artifact publication occurs.

## Requirements

- FR-001: Standalone `github.com/wstein/sheetmill` module, executable `smill`.
- FR-002: Public metadata uses approved description/topics; main is protected.
- FR-003: Pin Go 1.27.2 and Actions by verified commit SHAs.
- FR-004: CI checks format, dependency integrity, race tests, vet and bounded fuzzing.
- FR-005: Spec Kit Codex scaffolding and SheetMill constitution remain portable.
- FR-006: Explore three brand concepts without adopting a palette.

## Success criteria

Version and failure tests pass. Foundation files are present. CI/protection results
are recorded separately from local checks. Human review remains a merge gate.
No renderer, release binary or generated publication is included.
