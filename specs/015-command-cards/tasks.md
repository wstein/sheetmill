# Implementation tasks — 015

Implemented under the user’s 2026-10-14 T001–T011 handoff. See
[implementation evidence](implementation-evidence.md) for actual checks,
independent review and preserved baselines. PR/CI readiness is recorded on the PR;
no merge or deployment is authorized.

- [x] T001 Obtain a separate explicit implementation handoff; record scope in AGENTS.md and baseline existing parser/import/navi golden bytes and pinned identities.
- [x] T002 Write failing core validation and parser tests for `.command`, shared shapes/types, explicit IDs/references, empty/omitted mappings, nested positions, direct-IR fallback and unknown metadata preservation. Keep legacy-only diagnostics unchanged.
- [x] T003 Implement nonmutating shared command semantics, core/document validation and direct-IR checks; update schema and CMD001/CMD002 diagnostics only under that implementation handoff.
- [x] T004 Write paired legacy/new/dual/mixed renderer and CLI regression fixtures before migration integration. Hold source paths, IDs, titles, tool/tags, license, attribution, payload and options constant.
- [x] T005 For each successfully exported pair, assert complete byte-identical `.cheat` output and fixed expected per-entry pinned hashes/contexts. Test title/tag/path encoding, multiple cards, multiline payloads, source order and cross-sheet collisions; preserve existing hash vectors and prove NAV008 still rejects duplicate pinned identities before any writes.
- [x] T006 Implement navi consumption of shared semantics and legacy aliases once per card. Test equal duplicates, omitted/empty params, disjoint IDs, conflicting same-card maps, malformed legacy/new values and both markers; never partially merge same-card params or silently prefer one namespace.
- [x] T007 Test existing default/experimental shell profiles, legacy shell guard, wrong-profile skip/error behavior, placeholder diagnostics and strict builds. Prove old valid inputs gain no default deprecation warning; retain all existing payload/no-execution checks.
- [x] T008 Test and implement upstream shared-field/marker inert archival before core validation; preserve exact original values and markers with positioned import findings, unrelated metadata and existing import behavior. The new positioned import finding is IMP111.
- [x] T009 Prove companions retain authored old/new/mixed metadata while exported-card IDs/titles/contexts match; do not require sidecar byte equality or normalize away provenance.
- [x] T010 Update current format docs and 006/015 status only when implemented; retain a legacy example, publish migration and compatibility boundaries, and leave removal of aliases separately gated.
- [x] T011 Run appropriate Go tests/race/vet, no-execution lint, unchanged-golden checks and diff checks. Get independent Sol review/fixes/re-review, record actual results and create the separately authorized implementation PR. No navi execution, PDF, merge or deployment.

Dependencies: T001 first; failing T002 before T003. T004/T005 before T006; T007
covers migration integration. T008/T009 complete preservation before T010/T011.
The implementation handoff must deliver all migration and identity checks; a
field rename alone is not a complete first slice.
