# Analysis: 018 layout paths

## Coverage (FR/SC to tasks)

| Requirement | Tasks |
| --- | --- |
| FR-1, FR-2, FR-3 | T004, T005 (rule 7 in T002) |
| FR-4 canonical | T006 (rule 4), T002 |
| FR-5 sitemap, robots | T006 (rules 2 and 4, incl. manual sitemap), T002 |
| FR-6 mount, collisions, `demo*` prefix | T002 (rules 1, 3), T004, T005 |
| FR-7 verification | T002, T003, T004, T006, T006a |
| FR-8 determinism | T006 |
| FR-9 failure, rollback | T004 (`needs: build`, step order) |
| FR-10 cost | T008 (figures from the spec author's local measurement; not re-measured) |
| FR-11 security | T004 (permissions, pins), T006 (script/host assertions), T006a |
| FR-12 authorization | T001 (confirmed by the user 2026-10-16) |
| FR-13 documentation | T007 |
| SC-001 | T004, T006 |
| SC-002 | T002 (rule 7), T006 |
| SC-003 | T002, T006 |
| SC-004 | T006 |
| SC-005 | T006, T006a |
| SC-006 | T004 |
| SC-007 | T004 |

Unverified: the FR-10 sizes; the mobile layout of the header links; real-Hugo results
until `SHEETMILL_HUGO` tests run.

## Round 1 (reviewer on a746e19): NOT READY

Reviewer built manual and three demos with Hugo 0.167.0 and confirmed header-link
rendering, own-base URLs, page parity, byte-identical rebuild and file counts.

| # | Severity | Finding | Resolution |
| --- | --- | --- | --- |
| 1 | high | Rule "no self link" fails: logo and Library link home | Rule 7 limited to header-link elements |
| 2 | high | `http://` rule fails on the real manual (`127.0.0.1:1313` prose link); "//host" ambiguous | Rule 8: protocol-relative only; `http:` limited to src/link/action; regression test on real manual |
| 3 | medium | Authorization wording: T009 vs spec; AGENTS/012 incomplete | FR-12, T001, T009 fixed |
| 4 | medium | Integration test not run by `make test-hugo` | Package added to the target (T006) |
| 5 | medium | Script ran before llms/agent-guide copy | Script is last step; rule 1 checks top-level entries |
| 6 | medium | Marker rule imprecise/one-sided | Rule 5 token and regex based, both directions |
| 7 | low | Prefix rules ambiguous | Rule 3 states trailing slash and manual handling |
| 8 | low | FR-5 reason wrong | Reworded as a deliberate choice |
| 9 | low | FR-3 untestable | Restated as flag set and page set |
| 10 | low | Tool fragility | Double-quote assumption tested; `grep -E`/`sed`; `sh` invocation; `run:` parsed with mvdan.cc/sh |
| 11 | low | Identical titles | Variant titles distinct |

## Round 2 (reviewer on 445a69c): NOT READY

Facts re-checked against the code in this round: header links are `nav aria-label="Related
sites"` in `baseof.html` for editorial and `.topbar-link` anchors for the classic layouts,
hidden below 640px in `classic.css`; the Hugo `baseURL` placeholder `example.invalid`
is set in `hugo.go`; `pages.yml` is a single build run block; `internal/cli.Run` and
`docs/manual-pages.txt` exist; the `127.0.0.1:1313` link is in `docs/hugo.md`; `mvdan.cc/sh/v3`
is in `go.mod`; `make test-hugo` runs only `internal/render/hugo` today.

| # | Sev | Finding | Resolution |
| --- | --- | --- | --- |
| 1 | high | Rule 8 has no negative tree and no real-manual regression; T002 covers rules 1-7 only | T002 covers rules 1-8; T006 builds the real manual from `docs/manual-pages.txt` and runs the script; spec Testing says "real manual" |
| 2 | med | FR-7 "test fails on single-quoted/unquoted href/src" has no task | T006a template-text test in `internal/render/hugo` |
| 3 | med | Plan and FR-7 disagree on the script's tools | One list in both: `sh` with `grep -E`, `sed`, `find`, `sort`, `comm` |
| 4 | med | FR-12 calls the new paths part of the authorized demo | Reframed as a new 012 extension pending the user's confirmation; T001 records it as pending, T005 and T007 blocked; no authorization claimed |
| 5 | med | No constitution check; shell verifier in a Go-first repo | Plan constitution check; shell kept as the smaller change, justified in FR-7 and plan (test-only Go package) |
| 6 | med | No SCs, FR-to-task table, task FR tags, scenarios | SC-001..007, user scenarios, coverage table above, FR tags on tasks |
| 7 | med | Location and execution of the shell tests unspecified | `internal/pagesworkflow` via `os/exec` `sh`, skipped without `sh` (a skip is not a pass) |
| 8 | low | Rule 1 "exactly" vs subset; `demo*` slugs rejected | Rule 1 states the subset check; `demo` prefix reserved in FR-6 |
| 9 | low | Verification step vs run block unclear | Separate step `Verify the Pages tree`; T004 asserts order |
| 10 | low | Cross-layout links on mobile; no page names its layout | Accepted limitations section |
| 11 | low | `example.invalid` leak | Added to rule 4 (public docs and examples contain none, so the rule is safe) |

## Round 3 (reviewer on bad5011, Plan/opus, read-only): NOT READY

| # | Sev | Finding | Resolution |
| --- | --- | --- | --- |
| 1 | high | T004 fails on today's workflow while T005 is blocked: red branch, T006 blocked | T004 lands with T005; T006 depends on T003 only; header states both |
| 2 | med | The request authorization is not recorded in AGENTS.md | T001 adds the 018 line; spec marks the quote unverified until then |
| 3 | med | FR-11/SC-005 point at tasks that do not test them | T006 asserts one script, no inline script/style, no foreign host |
| 4 | low | Manual sitemap not checked | Added to rule 4 |
| 5 | low | "012 FR-7" dangling | FR-007 |
| 6 | low | Plan "reads and writes nothing" | Reworded |
| 7 | low | Plan step 4 mixes blocked and unblocked | Split |
| 8 | low | pages.yml header comment goes stale | In T004/T005 |
| 9 | low | llms.txt copy is not a separate step | FR-7 reworded |
| 10-12 | low | 017: cards.css missing from plan structure, T010 `[P]` vs chain, authorization claim | Fixed in 017, see ../017-freshness/analysis.md |

The same review confirmed 017 is consistent after round 3 apart from 10-12.

2026-10-16: the user confirmed the `/demo-compact/` and `/demo-cards/` paths (answer "1j"); this covers the paths only, not their activation, merge or deployment.

## Round 4 (final independent review): READY

| # | Sev | Finding | Resolution |
| --- | --- | --- | --- |
| 1 | med | Deploy gate misses single-quoted or unquoted `href`/`src`/`action`; T006a scans only `theme/layouts`, not HTML written by `hugo.go` (lines ~365, 385, 768) | Tag-anchored check `<[a-zA-Z][^>]* (href|src|action)=[^"]` in FR-7/T003 plus a T002 negative tree; extending T006a to the format strings noted as the unchosen alternative; 0 matches in current sources |
| 2 | low | T006 "one `<script src>` per tree" and foreign-host wording imprecise | Exactly one per HTML page (`<base>assets/app.<64hex>.js`); no absolute `src`/`<link href>` other than `https://wstein.github.io/` |
| 3 | low | T001 AGENTS.md line omits workflow edit and docs | Names T005 and T007, still excludes merge, deployment and workflow runs |
| 4 | low | Spec Authorization gives two scopes | One sentence listing spec, tests, script, workflow and docs edits, topic push, draft PR |
| 5 | low | T006 lacks a Hugo invocation helper | Own helper in `internal/pagesworkflow` mirroring the workflow flags and `HUGO_CACHEDIR` |
| 6 | low | Round 3 rows 10-12 point at nothing; confirmation not logged | Pointer to ../017-freshness/analysis.md; confirmation entry added |
| 7 | low | Script lacks SPDX header | T003 requires `# SPDX-License-Identifier: EUPL-1.2` |
