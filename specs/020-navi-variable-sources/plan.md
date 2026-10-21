# Plan - 020 navi variable sources

Status: Implemented on main (specification #60, core #63, export #64, all merged
2026-10-20); issues #61 and #62 remain open. Base: `origin/main` at the time of the branch.

## Architecture decisions

- The Markdown sheet stays the single source; `.cheat` is generated. No ingestion.
- The schema-1 IR is unchanged. Parameters stay in generic metadata; core reads them
  nonmutatingly (015). See [data-model](data-model.md).
- Core (`internal/validate/commands.go`, `CommandMetadata`) owns target-neutral shape
  rules for `choices`, `source` and `options` (`CMD001`), so direct-IR callers share
  them. `internal/render/navi` owns navi grammar limits (`NAV108`, `NAV109`) and the
  `$` line writer. The legacy parser (`parseCard`) accepts the same fields (`NAV006`).
- Pure functions, no I/O: `quoteChoices(profile, choices) string`,
  `variableLine(name, command, options) string`, a validator for navi text limits and
  a test-only model of navi's variable-line grammar (regex, `---` split, option
  words). The existing `quoteProblem` provides the source quote check.
- Entry identity is untouched: the prehash is still tags plus title.
- Trust: no flag, no gating (owner decision, 2026-10-20). The no-execution lint test
  stays as the scope boundary.
- Diagnostics continue the sequence: `NAV108`, `NAV109` (warnings); `CMD001`/`NAV006`
  extended. Documented in `docs/diagnostics.md` style.
- Hugo, Typst and other targets ignore the fields; their outputs stay unchanged.

## Evidence

[research.md](research.md) lists what source reading verified and what is
to-be-measured. No navi, shell or snippet runs in tests or CI.

## Verification

Failing-first tests per task, byte goldens per profile, existing goldens and hash
vectors unchanged, fuzz target for the line writer, mutation probes, source and
direct-IR negative tables, CLI tests for each `--shell`, schema consistency test.
No hard-coded tool or navi version in tests: the pinned navi evidence is read from
documentation only; the Go parser version is read from `go.mod` where a test needs it.

## Delivery sizing

Two small PRs, the second stacked on the first:

1. Core: shared validation, schema, import regression (T001-T005). Until PR 2 the
   navi build rejects the validated fields with NAV006 and the format pages are
   unchanged; the two PRs merge in order.
2. Export: writer, quoting, diagnostics, goldens, CLI, all docs (T006-T016).

Each PR needs checks green and an exact-head independent review before Ready;
merge, tags, releases and deployment remain with the owner.
