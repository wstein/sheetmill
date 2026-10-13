# Implementation Plan: SheetMill Studio — Local Live Preview

**Branch**: `codex/007-studio-spec` | **Date**: 2026-10-13 | **Spec**: [spec.md](spec.md)

**Input**: `specs/007-live-preview/spec.md`, including the user's narrowed first slice.

## Summary

**Implementation status (2026-10-13)**: the user authorized and the branch implements
T001–T018 plus MVP final checks only. Later stages remain design, not implementation.
See [verification](../../docs/e2e-verification.md).

Deliver one real `smill` reference through actual CLI validation/inspection/export,
pinned Hugo generation and an automated browser journey before adding Studio server
behavior. Reuse the existing Go compiler, semantic IR and Hugo renderer unchanged
unless a reproducing journey exposes a defect. Use a development-only Playwright
harness to verify keyboard search, card navigation, actual clipboard bytes,
light/dark behavior, page fit, no-JavaScript navigation and deterministic output. The complete scripted
reader journey runs unchanged in desktop and mobile viewports; only the separate
no-JavaScript context omits enhancement assertions.

The MVP ends after US1. No `serve` implementation, sibling sheets, new product API,
PDF, navi, deployment, screenshot baseline campaign or visual editor is required
for that gate. Later feature stages reuse the same harness for US2 and add the
read-only local `smill serve` session (US3–US5). This plan documents those stages
without authorizing or making them prerequisites for the MVP.

## Technical Context

**Language/Version**: Existing Go 1.27.2 production compiler/CLI. Node 24.21.0 LTS
and JavaScript ES modules for the isolated development test package only.

**Primary Dependencies**: Existing Cobra 1.10.2, Goldmark 1.8.6, YAML v3.0.5 and
Hugo 0.167.0. New development-only `@playwright/test` exactly 1.64.0, committed npm
lockfile and its bundled Chromium revision 1248 (156.0.8078.4). No new production
Go or frontend dependency for MVP. Native server stages use standard library first.

**Storage**: Committed independently authored sheet plus explicit test oracle;
owned temporary source/project/public trees; ignored local stage logs/screenshots/
traces outside public output. Later sessions retain bounded in-memory revision
state and owned generated snapshots, never edit source or Git state.

**Testing**: Existing Go tests/race/vet/fuzz; Playwright Chromium, one worker,
zero retries, desktop 1280×800, mobile 390×844 and a separate JavaScript-disabled
context. Subprocesses use fixed executable/argument arrays, never a shell.

**Target Platform**: Linux CI and macOS local development. Chromium only for MVP;
no claim of Safari/Firefox or Windows process-lifecycle coverage.

**Project Type**: Go CLI/compiler with a separate development-only browser harness;
later a native read-only local HTTP preview server using the same compiler.

**Performance Goals**: MVP correctness and bounded test execution, not speculative
production benchmarks. Later US3/US4 validate SC-002: 95/100 saved changes visible
within 1 second and 19/20 starts within 3 seconds for 100 sheets ≤100 KiB each;
SC-006 is a separate five-author usability exercise, not automated evidence.

**Constraints**: One `smill` sheet in MVP; no sibling runtime, snippet execution,
network resource fetches during tests, automatic tool downloads, distribution
builds or publication. Tool/browser installation is an explicit setup step, not
part of tests. Required E2E CI fails when prerequisites are absent; optional local
Go integration skipping does not count as an E2E pass.

**Scale/Scope**: One authored sheet and one complete reader journey in both scripted
desktop/mobile contexts plus a separate no-JavaScript context first. US2 adds two data cases; US3–US5 introduce
bounded local sessions. Source size limit remains the existing 8 MiB per file.
Later session ceilings: 100 selected files, 32 MiB aggregate source, 64 MiB total
generated snapshot bytes, 8 browser streams; exceeding a limit is explicit and
never truncates silently. Limits include retained last-valid state.

## Constitution Check

*Pre-research and post-design evaluation; no constitution amendment is required.*

| Principle | Design evidence | Result |
| --- | --- | --- |
| Determinism | Two independent complete public trees, exact file lists/bytes; fixed Hugo clock, evidence outside output | Pass |
| Canonical source | Real CLI consumes GFM+; no JavaScript compiler, source reserialization or schema change | Pass |
| Loss-aware imports | No new import adapter or import shortcut; raw metadata preserved; adversarial inputs separated from useful references | Pass |
| Inert input | Fixed test command allowlist; no commands derived from source; off-origin browser requests both blocked and recorded as failures | Pass |
| Focused Go design | Existing compiler reused; JavaScript restricted to development tests; later HTTP/watch work uses standard library | Pass |
| Independent product | `bdock`/`whr` are later source data, never installed runtime dependencies or invoked during journeys | Pass |
| Human publication control | This change contains planning documents only; future test executables are temporary verification tools, not distribution artifacts; no uploads/deploy/merge | Pass for planning |
| Review evidence | Separate actual checks from planned checks; independent review remains a merge expectation | Pass |

The original planning request was followed by explicit implementation authorization
for the MVP on 2026-10-13. That authorization selects T001–T018 and applicable final
checks only. Governance for later server implementation must reflect a subsequent
instruction without changing
Typst/navi or publication permissions. No automatic workflow may cross that boundary.

## Project Structure

### Documentation (this feature)

```text
specs/007-live-preview/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── checklists/requirements.md
├── contracts/e2e.md
├── contracts/serve.md
└── tasks.md
```

### Source Code (planned repository-relative paths)

```text
examples/sheets/smill.md                    # one authoritative MVP reference
tests/e2e/                                # development package (no runtime UI)
 ├── package.json / package-lock.json / .node-version
 ├── playwright.config.mjs
 ├── corpus.json / README.md
 ├── helpers/process.mjs
 ├── helpers/site.mjs
 ├── helpers/fixtures.mjs
 ├── helpers/network.mjs
 ├── harness.test.mjs
 ├── smill.spec.mjs
 ├── corpus.spec.mjs                       # later US2
 └── studio.spec.mjs                       # later US3–US5
internal/cli/serve.go                       # later only
internal/studio/
├── selection.go / session.go / compile.go  # later only
├── watch.go / http.go / security.go
└── web/client.js / web/status.css
.github/workflows/e2e.yml                  # proposed MVP CI, no artifact upload
```

**Structure Decision**: Keep the harness beside, not inside, production renderer
code. A Node fixture compiles `cmd/smill` once to an owned temporary test executable,
then directly invokes `version`, `check`, `inspect` and `build hugo` using arrays.
It invokes an explicitly supplied Hugo 0.167.0 executable and serves only an
allowlisted map of generated public files on `127.0.0.1:0`. It never uses the
existing preview server or requires implementing `smill serve` first.

## Phase 0 — Research decisions

See [research.md](research.md) for primary sources and alternatives. Resolved choices:
exact test toolchain pins; real subprocess entrypoint; in-process loopback fixture
server; independent content oracle; exact clipboard readback; explicit offline test
boundary; complete output comparison; one-worker isolated cleanup and local evidence.
No unresolved design clarification remains. Dependency installation, lockfile
creation and action commit/checksum verification are implementation tasks, not
operations performed by this planning invocation.

## Phase 1 — Design and contracts

[Data model](data-model.md) defines corpus cases, generated sites, journeys and
later session revisions. [E2E contract](contracts/e2e.md) is the MVP acceptance
interface. [Serve contract](contracts/serve.md) is explicitly later-stage design.
[Quickstart](quickstart.md) separates existing checks from proposed validation
commands, which do not exist until implementation.

MVP stages: validate pins → create owned workspace → compile temporary CLI → hash
source → check/inspect → export into two distinct new directories → run Hugo twice
→ compare complete output → start owned loopback server → browser assertions →
verify unchanged source → teardown. Every stage records bounded local evidence.

Later live preview uses content hashing/polling at 100 ms and 100 ms quiet-period
coalescing (500 ms maximum coalescing delay). One publishing subprocess runs at a
time with immutable revision snapshots; obsolete results cannot be committed.
Reuse the renderer for valid and retained last-valid per-sheet IR, mark stale
sheets explicitly and isolate invalid sheets from valid ones. Generated files
remain unchanged on disk; a serve-only trusted client script adds status chrome
when serving HTML. Ordinary export output and deterministic comparisons exclude
no published file. Session state is a separate channel, never canonical IR.

Post-design gates pass as above. Browser writing, rich-text packages, WebAssembly,
public hosting and later renderer integration remain outside this plan's execution
scope. Plan/tasks are reviewable design, not evidence that any feature is implemented.
