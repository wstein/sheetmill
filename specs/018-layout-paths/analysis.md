# Analysis: 018 layout paths

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
