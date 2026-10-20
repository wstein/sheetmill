# Validation guide - 020

Acceptance guide for the implementation; results are recorded on the PR, not here.
Never run navi, a shell or a sheet command; tests are in-process.

1. Baseline (T001): the existing navi goldens, migration fixtures, sidecar goldens and
   parser goldens must pass unchanged at the end.
2. Core: run `go test ./internal/validate -run 'TestCommandParam|TestSchemaMatchesVariableSourceRules'`
   and confirm positioned `CMD001` for each row of the negative table and the `1:1`
   anchor for direct IR.
3. Export: `go test ./internal/render/navi -run 'TestVariable|TestChoices|TestSource|TestQuote'`
   and the per-profile golden tests; review each golden by eye against the contract.
4. CLI: `go test ./internal/cli -run TestBuildNaviVariableSources`; check that an
   error writes no output directory and `--strict` fails on `NAV108`/`NAV109`.
5. Fuzz: `go test ./internal/render/navi -run '^$' -fuzz FuzzVariableLine -fuzztime 5s`.
6. Mutation probes (T014) and the whole suite: `make check`.
7. Identity: compare a sheet exported with and without `source`/`choices`; entry
   hashes and `NAV008` behavior are equal.
8. Optional, separately authorized: the to-be-measured items in
   [research.md](research.md).
