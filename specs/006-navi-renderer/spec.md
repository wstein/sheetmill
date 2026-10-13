# 006 — navi renderer

Created: 2026-10-13 · Revised 2026-10-14 · Status: implemented on `worktree-navi-export`,
awaiting independent review · Depends on: 002-gfm-parser and accepted 001–004

The normative contract is [docs/navi.md](../../docs/navi.md); this file records the
decisions and their rationale. Export is one-way text generation: navi, a shell and
exported commands are never run.

## Decisions (user, 2026-10-14)

- Output is `.cheat`, one file per sheet.
- A card is exported only with the `{.navi}` class **and** an explicit `{#id}`.
  Generated IDs change with titles, so they cannot anchor parameter metadata.
- Shell profiles: bash first, then zsh and fish, nushell as an optional profile.
  navi runs snippets through one configured shell and the fence label does not
  select it, so a build targets one profile and records it in the header.
- A card that cannot be exported is a **warning**: it is skipped, nothing is
  rewritten to make it pass. Metadata and configuration errors fail the build.
- No navi execution, including no pinned-navi parse check in CI.

## Changes from the first draft

- Parameters attach to card IDs (`cheatsheet.navi.cards.<id>.params.<name>`), not to
  bare placeholder names. Types are `string|int|path`; choice lists, defaults and
  dynamic `$ name: command` suggestions are deferred.
- Placeholder grammar is separate from shell syntax. `<name>` matching
  `[a-z][a-z0-9_]{0,31}` is a placeholder; redirections and heredocs are shell.
  Names navi would read as variables but the grammar rejects are `NAV103`.
- Marker rule corrected after reading upstream `parser.rs` (master): column-0
  `%`, `@`, `;`, `#`, `$` and fence lines are syntax even inside a fence; `@` was
  missing from the draft. Indented lines are safe.
- Per-sheet variable scope: one name declared differently in two cards is `NAV007`.
- Codes: errors `NAV001`, `NAV006`, `NAV007`; warnings `NAV101`–`NAV106`
  (replaces the draft's `NAV002`–`NAV005` errors and leading-dash warning).

## Requirements

- FR-001: Consume schema-1 IR directly; never reparse source or run snippets.
- FR-002: Deterministic bytes; stable order by sheet path then card order; no
  clocks, host paths or environment data.
- FR-003: New output directory only; validate everything before writing.
- FR-004: Header carries source, title, shell profile, license (or "not stated")
  and attribution, each collapsed to one line.
- FR-005: The renderer package imports no `os/exec`, `syscall`, `plugin` or network
  package (lint test).

## Verification

Table-driven negative tests, golden bytes, a pure Go model of navi's line grammar
recovering structure from the output, a fuzz target over command text, CLI tests.

## Remaining decisions

Static choice lists (with how they avoid executable providers), shared-context
inheritance, nushell raw-string quote analysis, whether a non-executing parse check
against a pinned navi exists, publication of generated cheats (gated).
