# Tasks - 020 navi variable sources

**Input**: [spec.md](spec.md), [plan.md](plan.md), [research.md](research.md),
[data-model.md](data-model.md), [contract](contracts/variable-sources.md),
[quickstart.md](quickstart.md).
**State**: implemented on main: specification merged in #60, PR 1 in #63, PR 2 in #64
(all merged 2026-10-20); issues #61 and #62 stay open for the owner to close. Merge, tag, release, deployment, PDF and navi/shell
execution are not covered.

Write each failing test first, observe the failure, then implement. No test
hard-codes a tool, navi or dependency version. Each task is one commit; no
task exceeds about 200 changed lines without a stated reason.

## PR 1 - core (T001-T005)

PR 1 validates, schema-documents and preserves the fields. Until PR 2 the navi
build rejected these fields with NAV006 and the format pages stayed unchanged (stated gap;
the PRs merged in order, so PR 2 (#64) closed it).

- [x] T001 Run the existing navi goldens, migration fixtures, sidecar goldens and parser goldens unchanged as the baseline; no new checksum file (the committed byte goldens are the baseline). Record the base SHA on the PR.
- [x] T002 Failing tests in `internal/validate/commands_test.go`: `TestCommandParamChoicesShape` (empty, duplicate, non-string, empty string, whitespace, controls, newline, U+2800, too long, too many), `TestCommandParamIntChoicesAreIntegerStrings`, `TestCommandParamSourceShape`, `TestCommandParamChoicesAndSourceExclusive`, `TestCommandParamOptionsShape` (types, ranges, unknown key including `map`, dependencies on `source`/`choices`, `multi` with `prevent_extra`), `TestCommandParamNewFieldsPositions` (nested key spans; `1:1` for direct IR).
- [x] T003 Implement the rules in `CommandMetadata` (`internal/validate/commands.go`) with `CMD001`; no IR change, no mutation of metadata.
- [x] T004 Failing then passing: `TestSchemaMatchesVariableSourceRules` in `internal/validate/variable_sources_schema_test.go` (modeled on `TestSchemaMatchesFreshnessTables`), then edit `schema/cheatsheet.schema.json` per the contract.
- [x] T005 Failing tests first: `TestFrontmatterAcceptsVariableSources` and `TestFrontmatterRejectsVariableSourceErrors` in `internal/frontmatter/commands_test.go` (source path through `Parse`). Characterization tests, expected to pass without code changes: `TestGoldenCommandCardsUnchanged` (parser goldens byte-identical) and `TestImportKeepsUpstreamVariableSourcesInert` in `internal/importer/reference/command_test.go` (upstream `choices` stays archived, IMP111); fix only if one fails.

**Checkpoint PR 1**: `make check` green; independent review of the exact head.

## PR 2 - export (T006-T016)

- [x] T006 Failing tests in `internal/render/navi/variables_test.go`: `TestChoicesQuotingPerProfile` (bash, zsh, fish, nu x plain, quote, backslash, `$`, `%`, backtick, leading dash, non-ASCII) and `TestQuoteChoicesRoundTrips` (a test-only unquoter per profile recovers the list); implement `quoteChoices`.
- [x] T007 Failing tests: `TestVariableLineGrammarModel` (pure Go model of navi's variable-line parser: name regex, first colon, `---` split, option words, `u8` ranges) and `TestVariableLineModelRejectsTrailingBackslash`; implement `variableLine` and the extended `param`/`parseCard` (legacy `NAV006` included).
- [x] T008 Failing tests with goldens `internal/render/navi/testdata/variable-sources/{bash,zsh,fish,nu}.cheat` copied from the contract: `TestVariableSourceGolden`, `TestChoicesLineInGolden`; wire the writer into `renderCard`.
- [x] T009 Failing tests: `TestSourceExportedVerbatim`, `TestSourceWithOptionsFixedOrder`, `TestVariableLinesFollowFirstUse`, `TestRepeatedPlaceholderOneLine`, `TestNoVariableLineWithoutSourceOrChoices`; make them pass.
- [x] T010 Failing then passing: `TestSourceRejectedSkipsCard` (table: `---`, navi token, trailing backslash, unterminated quote for bash/zsh/fish; nu has no quote check), `TestChoicesRejectedSkipsCard` (`---`, token), `TestOptionStringRejected` (`---`, `'`), `TestStrictFailsOnVariableSourceWarnings` (`NAV000`), `TestUnusedParamWithSourceStaysNAV104`. Implement `NAV108`/`NAV109` on card code-block spans.
- [x] T011 Failing then passing: `TestEveryDollarLineIsReferencedAndEntryHasNoBlankLine` (invariants over all goldens and a generated corpus), `TestSourceAndChoicesBothSetErrorsAtomically` (nothing written), `TestLegacyNamespaceAcceptsVariableSources`, `TestConflictingVariableDeclarationsAcrossNamespaces` (`NAV006`).
- [x] T012 Failing then passing: `TestEntryHashesUnchangedBySources` (equal pinned hash, context and tag line with and without sources; `NAV008` still rejects duplicates), `TestSidecarPreservesVariableSources`, `TestHugoIgnoresVariableSources`.
- [x] T013 `FuzzVariableLine` in `internal/render/navi/variables_fuzz_test.go`: accepted input yields text the grammar model parses to one variable with the same options; rejected input warns, never panics. `TestNoExecution` already forbids exec, syscall, plugin and network imports; confirm it covers the new files.
- [x] T014 Mutation probes, recorded in the PR (each fails at least one named test, then is reverted): drop `'\''` doubling; drop fish `\'` escaping; reorder choices; skip the `---` check in sources; skip the `---` check in option strings; skip the token check; emit a blank line before `$`; sort `$` lines by name; include `$` lines in the prehash; omit `NAV109`; accept source plus choices; accept `multi` with `prevent_extra`; coerce numeric choices.
- [x] T015 `TestBuildNaviVariableSources` in `internal/cli/navi_test.go` for each `--shell` (golden bytes, companion, exit codes, no output directory on error).
- [x] T016 Documentation (documentation text verified on main 99fc998 by grep; commits de721a3 and 004d103; #64 merged; the check/review steps below are recorded on #64, whose independent review was still "pending" in its body: unverified): `docs/format/navi-profile.md` (metadata table, placeholders, diagnostics, the "Defaults, choice lists ... not generated" and "`$` provider lines are never emitted" sentences, scope note that SheetMill writes text only), `docs/format/command-cards.md` (declarations table; remove "choices ... out of scope"), `docs/diagnostics.md` (`CMD001` row, `NAV108`, `NAV109`), `docs/cli.md`, and the stale statements in `specs/006-navi-renderer`: spec.md S5 ("no `$` lines are emitted", line ~54), plan.md ("choice lists without silent executable providers", line ~17), research.md ("no executable providers", line ~31, and "Escaping literal placeholders, choice lists and provider capabilities", line ~42), tasks.md T011. Then run `gofmt`, `go test ./...`, `go test -race ./...`, `go vet ./...`, the fuzz target and `git diff --check`; record exact revision and results on the PR; independent review of the exact head; open the PR. No navi execution, PDF, merge or deployment.

Dependencies: T001 first; T002 before T003; T004-T005 after T003. PR 2 starts after
PR 1's head is reviewed (stacked). T006 before T007; T007 before T008-T012; T013
after T008; T014 after T012; T015 after T010; T016 last.

## Later features (not in this feature)

Dependent variables (`$ b: cmd <a>`, `$a`), `@` shared contexts, further navi options
(`--query`, `--filter`, `--header`, `--preview`, `--preview-window`, `--fzf-overrides`,
`--expand`, `--map`), `.cheat` ingestion, choices in other renderers, and the measurements
listed in [research.md](research.md).
