# Analysis and review record — 017

## Coverage (FR/SC to tasks)

| Requirement | Tasks |
| --- | --- |
| FR-001 | T003, T004, T017 |
| FR-002, FR-003 | T003, T004, T017 |
| FR-004 | T005, T006 |
| FR-005 | T003, T018 |
| FR-006 | T007, T008 |
| FR-007 | T009, T011 |
| FR-008 | T010 (CSS-level only; the rendered 320px wrap is an unverified gap), T011 |
| FR-009 | T007, T008 (front matter and no `<lastmod>`) |
| FR-010 | T013, T014 |
| FR-011 | T005, T015 |
| FR-012 | T018 |
| FR-013 | T019 |
| FR-014 | T002, T016, T020 |
| SC-001 | T003, T006 |
| SC-002 | T012 |
| SC-003 | T002, T009 |
| SC-004 | T002, T015, T019 |
| SC-005 | T009 |
| SC-006 | T019 |

Task numbers in the round records below are as of the round; since round 3 the CSS
tests are T010 and the implementation T011 (T010a is gone, there is no gap).

## Unverified (not claimed)

Rendered wrapping at 320px and the desktop `index-fresh` row (no browser); upstream
correctness of demo versions (sheets' own notes only); real-Hugo results until
`SHEETMILL_HUGO` tests run.

## Round 1 (reviewed revision 07fcbdf5449e, Plan/opus, read-only)

Verdict: NOT READY (2 high, 7 medium, 10 low). Verified true: YAML timestamp strings,
location keys, unused VAL004/IMP112, navi companion, importer merge need, derived
asset names, demo dates and versions.

| # | Sev | Finding | Resolution |
| --- | --- | --- | --- |
| 1 | high | SC-003 byte identity impossible (fingerprinted CSS names); no HTML baseline | SC-003 restated with normalised hrefs; T002 stores normalised base hashes |
| 2 | high | Output list wrong (`sheets/<id>.json`, tag data, IR, hugo.json) | Contract and data model corrected; tag keys accepted, tag HTML shows none |
| 3 | med | CSS separator announced / runs together in plain text | Real aria-hidden separator span (FR-007) |
| 4 | med | Specificity loses to `.library-item p`; unscoped sibling rule | Scoped selectors, asserted by T010a |
| 5 | med | Classic row cannot wrap; cards header centring | `.index-fresh` full width; cards.css centring; 320px browser check stays a gap |
| 6 | med | Markup tests skip without Hugo; clock helper hard-coded | Two-layer T009, skip is not a pass, clock helper in T012 |
| 7 | med | Authorization not in repo | Assumption says T001 writes it first |
| 8 | med | Missing coverage table and gaps | This table; T007, T010a, T006, T019 extended |
| 9 | med | Navi from source does validate | FR-011, contract, T005 CLI test |
| 10 | low | Markup mismatch research/contract/US1.2 | Harmonised |
| 11 | low | Task order and dependencies | T010a before T010; T016 needs T005 |
| 12 | low | Baseline hashes only json | Hash all of `testdata/` plus navi export |
| 13 | low | agent-guide/llms.txt | Added to FR-012, T018 |
| 14 | low | Fuzz seeds | In T005 (`f.Add` seeds) |
| 15 | low | Performance budget | Noted in plan and T010 |
| 16 | low | Blank-looking fillers | FR-003 rejects fillers; limitation documented |
| 17 | low | Direct IR non-mapping cheatsheet | Stated in data model and T006 |
| 18 | low | Schema test, YAML 1.1 dates | Schema-vs-table test; quote advice |
| 19 | low | Circular "last updated" wording | Research reworded |

## Round 2 (reviewed revision e0adb0118e9d, Plan/opus, read-only)

Verdict: NOT READY (1 high, 2 medium, 7 low). Round-1 resolutions otherwise confirmed.

| # | Sev | Finding | Resolution |
| --- | --- | --- | --- |
| 1 | high | Classic `flex-basis:100%` ineffective without wrapping; cards column | `.classic .index-group li{flex-wrap:wrap}`, cards reset; desktop layout marked unverified |
| 2 | med | Classic header placement/selector contradictory | Exact insertion point after description; `.classic .sheet-fresh` |
| 3 | med | Spaces inside aria-hidden span | Spaces outside, `·` hidden |
| 4 | med | Baseline hash procedure unspecified | Fixture, helper, `-update-base`, recorded Options/Hugo, lifetime |
| 5-10 | low | Fuzz seeds, schema regex dialect, data-model wording, git log revision, test layers, clock pairing | Applied in tasks, data model, research |

## Round 3 (reviewed revision 0f0548c, independent review)

Verdict: READY (no critical or high; fix 1-4 before implementation). Facts re-checked
against the code in this round: `home.html` row order is `a`, `.index-desc`,
`.index-count`, `.index-license` and all 15 demo sheets carry a license; Studio calls
`hugo.Render` (`internal/studio/compile.go`); `hugo.go` writes `<loc>`-only sitemaps and
front matter `title`/`url`/`sheetid`; the editorial build links only `site.css`;
`git diff 05eba4d 189e872 -- internal/render/hugo/theme` is empty; `validateHeaderLink`
uses `unicode.IsPrint`.

| # | Sev | Finding | Resolution |
| --- | --- | --- | --- |
| 1 | med | `.index-fresh` "after the card count" would strand the license under `flex-basis:100%` | Last child, after `.index-license`: contract, plan, US1.2, T009(a) |
| 2 | med | T020 lacks `make test-studio`; Studio shares the templates | Contract changed-output list and T020 |
| 3 | med | FR-008 "wrap at 320px" is unverifiable | FR-008 reworded to CSS-level assertions (T010); 320px listed as unverified here and in the spec |
| 4 | med | `tar` card would read bsdtar's 3.5.3 as bare "Version" | Decision: omit `tool_version` for `tar` (research, FR-013, contract, T019); no qualified value, no tool name |
| 5 | med | Status line stale | Updated, round 3 recorded |
| 6 | low | Research #9 called spaces-outside the existing pattern | Reworded: deliberate difference from the provenance markup |
| 7 | low | Front matter date risk | Plan, data model and spec guard; T008 asserts front matter keys |
| 8 | low | `Co`/`Cn` accepted while header links reject | Recorded in FR-003 with the toolchain-determinism reason |
| 9 | low | Sitemap `<lastmod>` undecided | Decided no (FR-009, out of scope, T008) |
| 10 | low | "every build changes site, classic, cards css" | "each emitted stylesheet"; `app.js` unchanged |
| 11 | low | T011 missing, T010a before T010 | Renumbered: CSS tests T010, implementation T011 |
| 12 | low | Base `05eba4d` vs PR base `189e872` | Both named in SC-003 and T002 |
| 13 | low | Unknown `cheatsheet.*` keys untested | Added to T003 and T005 |
| 14 | low | "freshness" clashes with Studio preview staleness | FR-012 and T018 wording |

Follow-up from the 018 round-3 consistency check (reviewed bad5011): `cards.css` added to
the plan structure, `[P]` dropped from T010 (it follows T009), and the authorization
assumption marked as not yet recorded in the repo.

Final independent review (both specs READY), finding 8 (low): the contract placed `.library-fresh` "after the `Tool:` line" although no demo sheet sets `cheatsheet.tool` and FR-013 forbids adding it. Resolved: the contract states the template position (after the optional `.library-tool`, before `.library-description`, same with or without tool) and T009(a) pins it.
