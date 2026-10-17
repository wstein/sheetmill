# Analysis: 019 release workflow

## Coverage (FR/SC to tasks)

| Requirement | Tasks |
| --- | --- |
| FR-1 triggers | T004, T005 |
| FR-2 verify | T002, T003 (git checks), T004, T005 (API checks, order) |
| FR-3 build | T004, T005 |
| FR-4 smoke | T002, T003, T004, T005 |
| FR-5 archives | T002, T003, T004 |
| FR-6 publish | T002 (notes), T004, T005 |
| FR-7 security | T004 (mutations) |
| FR-8 tests | T002, T004, T006 |
| FR-9 pipeline | T001, T006 |
| FR-10 docs | T007 |
| SC-001..006 | T004 (SC-001, 005, 006), T002 (SC-002, 003, 004) |

Unverified until the first real tag: annotated-tag fetch, check-run API query, release
existence check, `ls-remote`, environment gate, `gh release create`, runner labels.
Not run by the author: any workflow, any tag operation.

## Review rounds

(recorded below)
