# Analysis and review record — 017

## Coverage (FR/SC to tasks)

| Requirement | Tasks |
| --- | --- |
| FR-001 | T003, T004, T017 |
| FR-002, FR-003 | T003, T004, T017 |
| FR-004 | T005, T006 |
| FR-005 | T003, T018 |
| FR-006 | T007, T008 |
| FR-007 | T009, T010 |
| FR-008 | T010a, T010 |
| FR-009 | T007 |
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
