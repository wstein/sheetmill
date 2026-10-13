# Tasks: SheetMill Studio — Local Live Preview

**Input**: `specs/007-live-preview/{spec,plan,research,data-model,quickstart}.md` and
`contracts/{e2e,serve}.md`. MVP tasks are implemented as marked; later tasks remain planned and unchecked.
**Tests**: Explicitly requested by the spec and repository test-first policy.
**Format**: `- [ ] Tnnn [P?] [USn?] Description with exact repository-relative paths`.

## Delivery boundary

MVP is **US1 only**: one `smill` sheet, real CLI → Hugo → complete automated browser
journey. Complete T001–T018 plus the MVP portions of T047/T048/T050, then stop for
review. No sibling binaries, `serve`, PDF, navi, visual editor or deployment.
Later task presence is not permission to execute those stages. US2 can be the next
increment after MVP; its lower priority places it later in this document, not behind
server dependencies. US3–US5 form one safe later server increment and must not delay MVP.

## Phase 1: Setup (development harness only)

- [X] T001 [P] Document the one-sheet execution boundary and fixed tool allowlist in `tests/e2e/README.md`, including offline setup versus test execution and no snippet execution/artifact uploads.
- [X] T002 Create private `tests/e2e/package.json`, committed `tests/e2e/package-lock.json` and `tests/e2e/.node-version` with Node `24.21.0` and exact `@playwright/test` `1.64.0`; add fixed `test` and `test:harness` scripts, no publish script, and record matching Chromium revision `1248` in `tests/e2e/README.md`.
- [X] T003 Configure `tests/e2e/playwright.config.mjs` for one worker, zero retries, desktop `1280×800`, mobile `390×844` and JavaScript-disabled projects; use `retain-on-failure` traces and `only-on-failure` screenshots outside public output under ignored `.local/e2e/`, adjusting `.gitignore` only if required.

## Phase 2: Foundational (MVP-only prerequisites)

**Purpose**: Own process/output boundaries without introducing any product server.

- [X] T004 Write failing Node tests in `tests/e2e/harness.test.mjs` for direct executable arrays, prerequisite/version failures, per-command deadlines, bounded logs, complete file-map comparison, unknown routes and cleanup after failed setup; follow the values in `data-model.md` rather than executing source snippets.
- [X] T005 [P] Implement `tests/e2e/helpers/process.mjs` to satisfy T004 using `shell:false`, fixed executable identities/argument arrays, per-command logs `≤1 MiB`, tool build deadline `120 s`, Hugo deadline `30 s`, CLI deadline `15 s` after compilation, redaction and termination/await of known children only.
- [X] T006 [P] Implement `tests/e2e/helpers/site.mjs` to satisfy T004: preload sorted regular output files, reject symlinks/non-regular entries, compare full path sets/bytes and serve a generated-file allowlist on `127.0.0.1:0` with MIME types/index routes; no directory listing, filesystem request traversal or existing-server reuse.

**Checkpoint**: Harness helpers are testable without authoring Studio or changing the compiler.

## Phase 3: US1 — One complete publishing journey (P1, MVP)

**Goal**: One authoritative reference and one real automated author-to-reader flow.
**Independent test**: `npm test -- --project=desktop` from `tests/e2e`, then `npm test`
for mobile/no-JavaScript checks, with explicit installed `SHEETMILL_HUGO`. No `serve`.

### Tests first

- [X] T007 [US1] Write the failing complete journey in `tests/e2e/smill.spec.mjs` against the real expected title/cards/search/copy/attribution and missing fixture; assert observable navigation and clipboard contents, never merely a “Copied” label.

### Implementation and integration

- [X] T008 [US1] Author `examples/sheets/smill.md` as a concise useful reference for currently implemented commands, verifying each example against CLI help/docs at a recorded full source revision; include stable IDs, code/table examples and EUPL attribution, without documenting future `serve` as available.
- [X] T009 [US1] Create `tests/e2e/corpus.json` with exactly one case `smill`, `input: examples/sheets/smill.md`, a nonempty exact title/query, expected stable card anchors, verified product identity/attribution and explicit expected multiline copy text that preserves whitespace/trailing newline; reject duplicate IDs, traversal, absolute paths and symlinks when loading it.
- [X] T010 [US1] Implement `tests/e2e/helpers/fixtures.mjs` with setup allowance `300 s` and reader-test allowance `120 s`: verify pins, resolve system-temp symlinks, build the real `cmd/smill` to an owned temporary test executable, run fixed version/check/inspect/export arguments, generate two sites using explicit pinned Hugo/fixed clock, compare all output bytes and source hashes, start the owned fixture server, and register failure-safe teardown from the beginning.
- [X] T011 [P] [US1] Implement `tests/e2e/helpers/network.mjs` to install a context-wide exact-origin request allowlist before navigation, set `serviceWorkers: block`, record every off-origin attempt and fail on any violation instead of silently masking it.
- [X] T012 [US1] Complete shared scripted reader assertions in `tests/e2e/smill.spec.mjs`: home → search → Tab focus → Enter card fragment → exact displayed code → sentinel-seeded actual clipboard readback, light/dark applied state plus reload persistence, stable card links, visible provenance and no horizontal page overflow; use semantic locators/condition-based assertions without clipboard mocks or sleeps.
- [X] T013 [US1] Run the complete T012 scripted journey at both desktop `1280×800` and mobile `390×844` in `tests/e2e/smill.spec.mjs`, including search/keyboard links, exact copy, appearance persistence and zero page overflow in both; add a separate `javaScriptEnabled:false` context whose only exemptions are enhancement controls, still requiring ordinary content/navigation; keep the same single `smill` corpus case.
- [X] T014 [US1] Add `tests/e2e/inert.spec.mjs` using temporary adversarial variants of the same source (not another real-product corpus entry) to prove raw HTML/shortcodes/snippets stay inert, external image attempts are not fetched, source hashes remain unchanged and the fixed subprocess log contains no promoted example command.
- [X] T015 [US1] Extend `tests/e2e/harness.test.mjs` with failing setup/timeout/interruption fault-injection cases, then correct `tests/e2e/helpers/fixtures.mjs` teardown so contexts/listener/known children always close, owned temp data is removed and separate failure evidence survives; no broad process kill or deletion of user paths.
- [X] T016 [US1] Add `.github/workflows/e2e.yml` with contents-read permission, existing pinned checkout/setup-go conventions, verified SHA-pinned Node setup for `24.21.0`, explicit `npm ci`/matching Chromium setup and checksum-pinned Hugo `0.167.0`; run helper tests and all browser contexts as required tests, with no artifact upload, release, deployment or permissive missing-tool skip.
- [X] T017 [US1] Run the one-sheet journey twice from independent owned roots with prerequisites already installed; require identical public file lists/bytes, unchanged source, zero off-origin attempts and all desktop/mobile/no-JavaScript assertions, recording actual environment/results and local evidence locations in `docs/e2e-verification.md`.
- [X] T018 [US1] Reconcile actual first-slice commands/evidence in `tests/e2e/README.md`, `docs/e2e-verification.md` and `specs/007-live-preview/quickstart.md`; explicitly leave `serve`, sibling sheets, live latency and five-author usability results unimplemented/unmeasured, and enforce the MVP stop before later phases.

**MVP STOP**: One `smill` sheet and one complete automated GFM → Hugo → browser journey.
Apply only MVP verification/review portions of final tasks. Later phases do not block
this deliverable and must not be silently included in its implementation.

## Phase 4: US3 — Local saved-file live preview (P1, later)

**Goal**: Read-only `smill serve`, after the MVP gate, with selected-file updates.
**Independent test**: Start on a copied source, save externally, observe new published
content without restarting. Release this increment only with US4/US5 acceptance.

- [ ] T019 [US3] Write failing lifecycle/selection tests in `internal/studio/selection_test.go`, `internal/studio/session_test.go` and `internal/cli/serve_test.go` for single-file/directory/default selection, invalid startup source, readiness, explicit port conflict, atomic save and shutdown; include the bounds in T020.
- [ ] T020 [US3] Implement `internal/studio/selection.go` with explicit regular-file or direct-directory selection, retained `os.Root` handles and per-read type checks: `100 selected files`, `8 MiB per file`, `32 MiB aggregate source bytes`; diagnose exceeded bounds without silent truncation, and a single-file selection must not expose siblings.
- [ ] T021 [US3] Implement bounded session/snapshot ownership in `internal/studio/session.go`: immutable saved revisions, owned temporary roots, `64 MiB generated bytes total, including retained last-valid state`, one in-flight publisher and at most one queued newest revision; no source/Git writes.
- [ ] T022 [US3] Implement `internal/studio/compile.go` using the existing Go parser/IR/renderer and explicitly pinned Hugo with fixed arguments/controlled environment; preserve independently valid sheets and retain only explicitly stale last-valid per-sheet IR, with no source-defined command/config or JS compiler.
- [ ] T023 [US3] Implement `internal/studio/watch.go` with content hashing every `100 ms`, `100 ms` quiet coalescing and maximum `500 ms` postponement; handle direct additions/renames/removals and atomic replacement, revalidating containment/symlink boundaries on every read.
- [ ] T024 [US3] Implement `internal/studio/http.go` generated-file-map serving, readiness and private state/event routing behind the US5 security layer; inject only trusted serve-specific client assets into HTML responses, leaving generated source/public snapshot bytes unchanged.
- [ ] T025 [US3] Add Cobra integration in `internal/cli/serve.go` and register it in `internal/cli/cli.go` per `contracts/serve.md`: optional explicit input, `--open`, loopback-only `--port`, installed pinned `--hugo`, actionable startup errors and fixed-argument browser launch with address fallback.
- [ ] T026 [US3] Complete owned-resource cleanup tests and implementation in `internal/studio/session_test.go` and `internal/studio/session.go`: `5-second graceful shutdown`, cancellation of known owned Hugo child, bounded escalation, closed listener/root handles/streams, and removal only of owned temporary state.
- [ ] T027 [US3] Extend `tests/e2e/studio.spec.mjs` using the existing process/browser helpers for externally saved valid content, selection retention, additions/removals, atomic replacement and process shutdown; require US4/US5 tests before exposing the server increment as complete.

## Phase 5: US4 — Diagnostics and freshness (P1, later)

**Goal**: Latest-save correctness, clear errors and recovery without restart.
**Independent test**: Invalid → repaired source, rapidly reordered revisions and
forced connection loss show truthful state and eventually the newest source only.

- [ ] T028 [US4] Write failing state-order tests in `internal/studio/state_test.go` for `pending | current | invalid | removed | failed`, stale last-valid output, invalid startup with valid peers, duplicate route origins and older completions finishing last.
- [ ] T029 [US4] Implement revision/state transitions in `internal/studio/state.go`, preserving compiler filename/line/column/code/message, monotonic sequence identity and a latest-only publication check; removed or failed sheets cannot appear silently current.
- [ ] T030 [US4] Add state serialization and SSE notifications/heartbeats in `internal/studio/http.go` with `8 browser streams`, `10-second` heartbeats, full-state fetch on reconnect and no source paths/credentials outside selected relative identities; reject excess streams explicitly.
- [ ] T031 [US4] Implement trusted status chrome in `internal/studio/web/client.js` and `internal/studio/web/status.css`: pending/invalid/last-valid/disconnected labels, source-relative diagnostics via `textContent`, latest revision check, reconnect/full-state refresh and stable fragment retention; no editor/write controls or PDF/navi placeholders.
- [ ] T032 [US4] Add failing rapid-save/disconnect browser cases before completing fixes in `tests/e2e/studio.spec.mjs`; require every stale/disconnected state visible and latest saved content/diagnostics after reconnection, with source hashes changed only by the test's explicit external-save operation.
- [ ] T033 [US4] Add controlled `100` rapid-save sequences and error/recovery checks in `internal/studio/state_test.go` and `tests/e2e/studio.spec.mjs`, ensuring all final results match the newest saved revision; do not depend on arbitrary browser sleeps for assertions.
- [ ] T034 [US4] Extend `docs/e2e-verification.md` with actual diagnostic/freshness/connection results, distinguishing these later server results from the static first-slice evidence and leaving performance/usability claims unasserted until T049.

## Phase 6: US5 — Local privacy and inert source (P1, later)

**Goal**: Browser-accessible read-only preview cannot widen selected-file access.
**Independent test**: Unauthorized requests, source variants and boundary races are
refused/inert while an authorized local preview remains usable.

- [ ] T035 [US5] Write failing `internal/studio/security_test.go` cases for cross/null Origin, forged Host/forwarded host, cross-site GET reads, wrong/replayed bootstrap secret, unauthenticated assets/events, forbidden methods and cross-origin embedding; do not assume GET is protected by CSRF middleware alone.
- [ ] T036 [US5] Implement `internal/studio/security.go` exact assigned Host/origin/Fetch Metadata enforcement, loopback-only startup and a cryptographic `32-byte one-use` bootstrap secret with random HttpOnly SameSite=Strict session cookie; consume bootstrap, redirect to a clean local URL and redact credentials as specified in `contracts/serve.md`.
- [ ] T037 [US5] Apply security middleware and local-only CSP/no-referrer/nosniff/no-store headers to every private/generated/state/event route in `internal/studio/http.go`; allow only GET/HEAD, 405 other methods, no permissive CORS or browser project-write/run endpoints, and no raw filesystem handler.
- [ ] T038 [US5] Extend `internal/studio/selection_test.go` with failing traversal/encoding/symlink-replacement/oversized aggregate cases, then ensure `internal/studio/selection.go` rejects them throughout the session without exposing unselected files or silently dropping inputs.
- [ ] T039 [US5] Extend `tests/e2e/studio.spec.mjs` with authenticated-origin access, unrelated origin/host requests, inert raw HTML/shortcodes and resource-fetch guards; disable tracing entirely for authenticated Studio contexts, retain failure screenshots/sanitized logs without bootstrap addresses, compare selected source bytes and require no unselected content or credential leakage in local evidence.
- [ ] T040 [US5] Validate publisher argument/environment allowlists and deadlines in `internal/studio/compile_test.go`, proving snippets/metadata cannot select executables, arguments, remote resources or user config; ensure only owned publishing children are cancelled.
- [ ] T041 [US5] Document the tested threat boundary, rejected-access evidence and read-only behavior in `docs/studio-security.md` and `docs/e2e-verification.md`; explicitly exclude hostile same-user software/OS compromise and any implied public-server support.

## Phase 7: US2 — Additional real-product inputs (P2, optional next after MVP)

**Goal**: Add BundleDock and WorkHarbor references to the same harness, no new runtime.
**Independent test**: `SHEETMILL_E2E_CORPUS=all npm test` passes without installed
`bdock`/`whr`, proving representative cards and cross-links for all selected sheets.

- [ ] T042 [US2] Record the authoritative full source revisions or released versions and verified command inventories for BundleDock/WorkHarbor in `docs/e2e-corpus.md`; do not assume a similarly named checkout or installed `whr` is authoritative, and perform no sibling repository edits.
- [ ] T043 [P] [US2] Author `examples/sheets/bdock.md` against its T042 identity with stable IDs, verified command/flag examples, provenance and at least one meaningful cross-reference; original content only and no runtime invocation from the journey.
- [ ] T044 [P] [US2] Author `examples/sheets/whr.md` against its T042 identity with stable IDs, verified workflow examples, provenance and at least one meaningful cross-reference; original content only and no accounts/containers/services in the journey.
- [ ] T045 [US2] Add failing parameterized cases in `tests/e2e/corpus.spec.mjs`, then extend `tests/e2e/corpus.json` and selection in `tests/e2e/helpers/fixtures.mjs` for explicit `SHEETMILL_E2E_CORPUS=all`, preserving default MVP `smill` only and the same shared browser assertions/oracle rules.
- [ ] T046 [US2] Run all selected references with sibling executables absent from PATH and verify cross-sheet/card links, search, attribution and exact copy using the same harness; record actual results in `docs/e2e-corpus.md` and `docs/e2e-verification.md` without executing displayed examples.

## Phase 8: Polish and cross-cutting checks (apply to the selected increment only)

- [X] T047 Run required `gofmt`, Go tests/race/vet/module checks, helper and browser tests plus `git diff --check`; record commands/results/gaps in `docs/e2e-verification.md` for only the delivered slice, with no optional E2E skip presented as a pass.
- [X] T048 Reconcile `docs/cli.md`, `docs/hugo.md`, `tests/e2e/README.md` and `specs/007-live-preview/quickstart.md` with actually implemented commands; document MVP E2E first and add later `serve` guidance only after its increment passes.
- [ ] T049 For the later server increment only, record the workstation and measure SC-002/SC-003/SC-006 in `docs/studio-verification.md`: `95 of 100` edits within `1 second`, `19 of 20` starts within `3 seconds` for `100` sheets `≤100 KiB` each, `100` latest-save sequences correct, and `four of five` authors complete recovery workflow within `five minutes`; keep these out of MVP completion claims.
- [X] T050 Obtain independent review of the delivered slice, fix/re-review findings with focused Conventional Commits, and record exact reviewed revision/check evidence in `docs/e2e-verification.md`; create only an authorized reviewable topic PR, with no merge, deployment, release or artifact upload.

## Dependencies & Execution Order

- Setup T001–T003 precedes foundational T004–T006. T004 tests must fail before
  T005/T006 implementation; those helpers can then be developed in parallel.
- US1 T007–T018 depends only on those MVP helpers. Within US1: T008 establishes the
  source IDs for T009; T010 requires T005/T006/T009; T011 can run beside T010;
  T012/T013/T014 require fixtures/network controls. T016 follows passing local
  assertions; T017/T018 close the evidence gate. Then MVP-specific T047/T048/T050.
- **US1 → stop/review → US2 or later server increment**. US2 requires US1 and pinned
  sibling identities only; it has no server dependency and is the recommended next
  content increment. T043/T044 may run in parallel after T042.
- US3 requires US1, not US2. US4 requires session/compile state from US3; US5 security
  tests/design can proceed alongside US3, but HTTP exposure and final server acceptance
  require US5 middleware and US4 truthful freshness. **Do not ship unsecured US3**.
- US3–US5 complete as one later safe preview increment before T049 or a usability
  claim. No stage depends on Typst/navi or deployment. The later priorities indicate
  user value, not permission to execute every unchecked task in one run.

## Parallel execution examples

- Setup: T001 documentation alongside T002 package setup; T003 follows tool pins.
- US1: T010 fixture orchestration alongside T011 network guard after foundational
  tests; browser tasks remain ordered because they share `smill.spec.mjs`.
- US3: T019 selection tests and session/CLI test cases can be split by their named
  files; T023 watcher work may proceed after T020/T021 while publisher work T022 runs.
- US4: State tests/implementation precede event integration; status CSS can be drafted
  separately from state code after the status contract is stable. Do not parallelize
  concurrent edits to `http.go` or `studio.spec.mjs`.
- US5: T035 security tests and T038 containment tests are different files; middleware
  and HTTP integration remain ordered. Process-boundary tests T040 are independent
  once `compile.go` exists.
- US2: T043 and T044 are independent sheet files after the two inventories are pinned.

## Implementation strategy

Deliver the smallest runnable increment. Write behavior tests before production/helper
changes and observe the intended failure. Commit focused changes after logical groups;
update task checkboxes only for verified work. The first implementation request should
select T001–T018 and MVP-only final checks explicitly. Later tasks must stay unchecked.

Install dependencies/tools/browser prerequisites explicitly; then run tests offline.
Keep tool execution fixed and content inert. Preserve the same Go compiler/IR/Hugo
path and use manual browser QA only as complementary evidence, not the automated gate.
MVP implementation was authorized on 2026-10-13. The private test package and
matching browser were provisioned explicitly; tests themselves download nothing.
No product server or later runtime feature was implemented. Independent source
re-review is clear; final documentation is reviewed at handoff. See
`docs/e2e-verification.md`.
