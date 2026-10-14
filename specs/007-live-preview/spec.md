# Feature Specification: SheetMill Studio — Local Live Preview

**Feature Branch**: `codex/007-studio-spec`

**Created**: 2026-10-13

**Status**: Read-only preview and corpus increments implemented; automated T049
checks implemented; five-author SC-006 usability study pending. See
[plan](plan.md), [tasks](tasks.md) and [verification](../../docs/studio-verification.md).

**Input**: User description: "SheetMill Studio — WYSIWYG, Live Preview & Playground. Start with smill serve, introduce smill playground next, and add an optional public online playground afterward. Reuse the same compiler across entry points. First slice: one smill sheet and one complete automated browser journey; then add bdock and whr to the same harness. Keep this milestone focused on GFM → Hugo → browser. Example commands remain displayed content."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Prove one real reference from source to browser (Priority: P1)

An author wants a real, useful `smill` reference whose source can be checked,
compiled and read in a browser. A maintainer wants the complete journey verified
automatically so later Studio work builds on observed behavior.

**Why this priority**: This is the user's explicitly selected first slice. It
proves the existing author-to-reader workflow before adding a long-running preview
session or browser editing.

**Independent Test**: Start with a committed, independently authored `smill` sheet
in a clean temporary workspace. Invoke the real command-line entry point to check,
inspect and export it; generate the website with the pinned publishing tool; serve
it locally; complete the browser journey without executing any displayed snippet.
This slice delivers value independently of the live-preview stories below.

**Acceptance Scenarios**:

1. **Given** the `smill` sheet and the documented tool versions, **When** the
   automated journey runs from source, **Then** validation succeeds, inspection
   identifies the expected sections/cards, and the generated website contains
   the expected reference title, stable card links and source attribution.
2. **Given** the generated library, **When** a reader searches for a known command,
   focuses a result with the keyboard and activates it, **Then** the browser reaches
   the correct card, and its displayed command matches the source.
3. **Given** a card containing a code block, **When** the reader uses its copy
   control, **Then** the actual copied text equals the source snippet, including
   whitespace; a success label alone does not establish correctness.
4. **Given** desktop and narrow browser windows, **When** the reader browses,
   switches light/dark appearance and follows card links, **Then** content remains
   readable, the selected appearance is applied and no horizontal page overflow
   occurs. With scripting disabled, reference navigation and content still work.
5. **Given** any command shown inside the sheet, **When** checking, rendering,
   browsing, searching or copying it, **Then** it remains content and is never
   executed by the test or application. Only explicitly enumerated test-tool
   invocations run.
6. **Given** the same source and pinned tools in two separate temporary workspaces,
   **When** each website is generated, **Then** corresponding published file bytes
   are identical; transient logs and browser traces are excluded from comparison.
7. **Given** a failed journey, **When** the maintainer investigates, **Then** the
   local test report identifies the failed stage and includes relevant command
   output plus browser evidence when the browser stage was reached.

---

### User Story 2 - Extend the same journey to sibling-product references (Priority: P2)

An author wants `bdock` and `whr` references to benefit from the same checks without
creating a separate testing framework for each product.

**Why this priority**: Real additional inputs expose content and layout gaps after
one complete journey works. They do not introduce product-runtime dependencies.

**Independent Test**: After Story 1 passes, add the BundleDock and WorkHarbor sheets
as inputs to the existing harness and run the same checks against each reference.

**Acceptance Scenarios**:

1. **Given** an authoritative product revision or version for each reference,
   **When** its sheet is added, **Then** the provenance records that identity and
   every documented command/flag is verified against that version's documentation
   or read-only help before becoming an accepted fixture.
2. **Given** all three sheets, **When** the shared journey runs, **Then** each is
   discoverable by title/search, its representative cards pass the same reader
   assertions, and cross-sheet links reach the intended cards.
3. **Given** an environment without installed `bdock` or `whr`, **When** the
   SheetMill journey runs, **Then** it still passes; their binaries, containers,
   accounts and services are not needed or invoked.

---

### User Story 3 - Follow saved source changes in a local preview (Priority: P1)

An author edits a selected sheet in Helix or another editor and keeps SheetMill
Studio open beside it. Saving the file updates the browser without manually
repeating export/build commands.

**Why this priority**: This is the first Studio capability after the source-to-
browser gate. It provides immediate authoring value without a visual editor.

**Independent Test**: Start `smill serve <file-or-directory>`, open its printed local
address, save a visible card change in an external editor, and observe the updated
preview and revision state without reloading the browser manually.

**Acceptance Scenarios**:

1. **Given** a regular selected `.md` file or one-level directory, **When** the author
   starts `smill serve`, **Then** the terminal reports the local address and selected
   scope, and the browser shows the available references or actionable source
   diagnostics. Source files, Git state and existing export directories are unchanged.
2. **Given** a running session, **When** a selected file is saved or atomically
   replaced with valid content, **Then** the preview updates to that saved revision,
   retaining the current sheet/card location where its identifier still exists.
3. **Given** a selected directory, **When** an eligible file is added, renamed or
   removed directly inside it, **Then** the visible reference list updates; nested
   files and files outside the selected scope remain inaccessible. A selected
   individual file does not grant access to its sibling files.
4. **Given** `smill serve --open`, **When** startup succeeds, **Then** the local
   preview opens in the browser. Without `--open`, startup prints an address without
   opening a browser. A browser-launch failure leaves a usable printed address.
5. **Given** a healthy session, **When** the author stops it, **Then** the process
   exits promptly, releases its listener and stops background work without changing
   project files. A port conflict produces a clear startup error.

---

### User Story 4 - Understand errors and preview freshness (Priority: P1)

An author needs to distinguish current valid output from an earlier successful
preview while repairing a source error or waiting for a new revision.

**Why this priority**: Silent stale output would make authoring misleading and hide
compiler errors; a running session must remain useful while the source is invalid.

**Independent Test**: Introduce an invalid edit, repair it, and then save several
revisions rapidly. Verify diagnostics, freshness labels and eventual latest output.

**Acceptance Scenarios**:

1. **Given** a previously valid preview, **When** the next saved revision is invalid,
   **Then** the browser shows the originating sheet, source line/column, diagnostic
   code and message; retained earlier output is explicitly labeled as the last valid
   revision, never as current successful output.
2. **Given** an invalid sheet at startup, **When** the author opens Studio, **Then**
   its diagnostics are visible even though it has no successful preview; other valid
   selected sheets remain readable. Repairing the sheet recovers without restarting.
3. **Given** rapid saves whose processing completes out of order, **When** the work
   settles, **Then** the latest saved revision alone determines the current preview
   and diagnostics; older results cannot replace it.
4. **Given** a preview connection that drops, **When** the browser loses contact,
   **Then** it labels the displayed output disconnected/stale. Reconnection reloads
   the current state without silently claiming freshness or losing sheet selection.
5. **Given** an unavailable output format, **When** the reader uses this milestone,
   **Then** only web output is offered as supported; no approximate PDF/navi preview
   is presented as compiler output.

---

### User Story 5 - Keep a local project private and inert (Priority: P1)

An author wants to preview a local reference without exposing unrelated files or
allowing another website or source snippet to operate on the machine.

**Why this priority**: A long-running browser-accessible session has different
boundaries from an offline compiler and must not widen project access implicitly.

**Independent Test**: Attempt access from an unrelated browser origin, from a
non-local network client and via paths outside the selected scope. Include inert
HTML, shortcodes and shell examples in dedicated adversarial fixtures.

**Acceptance Scenarios**:

1. **Given** normal startup, **When** another machine attempts to connect,
   **Then** it cannot access the session; no public/network binding option is provided
   by this milestone.
2. **Given** an unrelated website, **When** it attempts to read preview/source data
   or initiate operations in the local session, **Then** access is refused. Direct
   navigation to the CLI-issued local address remains usable.
3. **Given** unselected files (including hidden files), nested, symbolic-link or parent-relative paths,
   **When** the browser requests their contents, **Then** no data outside the explicit
   selection is disclosed. Unrelated repository files and secrets are not served.
4. **Given** source containing raw HTML, shortcodes, scripts, external image URLs
   or shell commands, **When** it is previewed, **Then** existing escaping/warning
   behavior is retained, no snippet runs and no linked resource is fetched
   automatically. Explicit user-followed ordinary links remain ordinary navigation.
5. **Given** an attempt to create, edit, delete or run a project file from Studio,
   **When** it is made in this milestone, **Then** it is rejected: the browser
   experience is read-only and does not provide terminal, Git or build-command access.

### Edge Cases

- Empty selection, nonexistent input, unreadable source or an occupied port:
  actionable startup error; no silent broader discovery or public binding.
- Invalid frontmatter/attributes, oversized or invalid UTF-8 source:
  existing compiler diagnostics; no browser/editor normalization of source bytes.
- Duplicate sheet routes, explicit duplicate IDs or unsupported URL schemes:
  originating-sheet diagnostics rather than merged cards or overwritten output.
- Temporary disappearance during an editor's atomic save:
  recover at the selected path; do not expand access to neighboring paths.
- A deleted selected file or removed card anchor:
  mark the unavailable reference clearly and provide navigation to remaining sheets.
- Connection loss, processing failure or unavailable local publishing tool:
  visible state/error; do not label an older or approximate preview current.
- Directory entries replaced by symlinks while the session is running:
  stop reading those entries and report the rejected boundary.
- Rapid edits followed by a slower older render:
  latest saved revision wins; no obsolete diagnostics or output overwrite it.
- File content including literal attribute syntax, raw blocks or imported metadata:
  retain canonical bytes and the existing compiler's preservation/loss reporting.
- Browser/test evidence containing session credentials or project content:
  redact credentials; keep evidence local by default; publish nothing automatically.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The first delivered slice MUST be one authoritative `smill` sheet and
  one automated complete GFM → Hugo → browser journey. Checks MUST invoke the actual
  CLI entry point, verify reader behavior and compare actual generated output, not
  stop at component tests or a static screenshot. Acceptance: Story 1.
- **FR-002**: The same harness MUST accept `bdock` and `whr` as later content inputs
  without executing their examples or depending on their product runtimes. Each
  sheet MUST identify its verified product version/revision and attribution.
  Acceptance: Story 2.
- **FR-003**: `smill serve [file-or-directory] [--open]` MUST create a local, read-only
  live-preview session for explicitly selected regular `.md` sources. Omitted input
  means the current directory's direct eligible files, not recursive discovery or
  an inferred project manifest. Acceptance: Story 3.1, 3.3–3.4 and Edge Cases.
- **FR-004**: The Studio preview MUST use the authoritative SheetMill compiler and
  the existing Hugo publishing path. Validation, metadata, card order, layout hints,
  escaping, source attribution and stable links MUST agree with ordinary CLI output
  for identical source/configuration/tool versions. A separate approximate editor
  compiler is prohibited. Acceptance: Stories 1, 3 and 5.4.
- **FR-005**: Saved changes and direct eligible directory-entry changes MUST refresh
  the browser automatically, without rewriting source, changing Git state or
  overwriting user export directories. Browser editing and autosave are excluded.
  Acceptance: Story 3.1–3.3 and 5.5.
- **FR-006**: Every displayed preview and diagnostic set MUST have an unambiguous
  revision/freshness state. Invalid updates, disconnects and failures MUST remain
  visible; only the latest saved revision may become current. Acceptance: Story 4.
- **FR-007**: Source errors MUST retain their filename, line, column, code and
  actionable message. Invalid sheets MUST be repairable without restarting the
  session, and independently valid selected sheets MUST remain usable.
  Acceptance: Story 4.1–4.2.
- **FR-008**: Sessions MUST be accessible only from the local machine, reject
  unrelated website access, enforce the selected filesystem scope throughout their
  lifetime, and never automatically fetch linked resources or execute source
  snippets. No browser-facing write/run capability is permitted. Acceptance: Story 5.
- **FR-009**: Previewing MUST work without Internet access once documented local
  prerequisites are installed. Startup and normal updates MUST NOT download tools,
  themes, fonts or other resources. Acceptance: Stories 1 and 3 run offline.
- **FR-010**: The reader experience MUST retain keyboard search/navigation, exact
  snippet copying, responsive layout, labeled freshness and readable light/dark
  presentation. Basic reference content/navigation MUST work without browser
  scripting; live refresh/search may require it. Acceptance: Story 1.2–1.4 and 4.4.
- **FR-011**: Startup conflicts, unavailable prerequisites and shutdown MUST have
  actionable terminal outcomes, with browser status when a session already exists.
  Shutdown MUST release local resources. Acceptance: Story 3.4–3.5 and Edge Cases.
- **FR-012**: This milestone MUST offer web output only. PDF/Typst, navi, visual
  editing, disposable/public playgrounds, collaboration, deployment and publication
  are deferred. Unavailable formats MUST NOT be simulated or silently substituted.
  Acceptance: Story 4.5 and the stated scope below.

### Key Entities *(include if feature involves data)*

- **Source selection**: Explicit regular Markdown file or direct directory entries;
  defines the maximum readable scope and never grants general repository access.
- **Reference revision**: A selected sheet's saved content identity and diagnostic
  state; determines whether preview output is current, pending, invalid or absent.
- **Preview session**: A local process, selected sources, authorized browser access,
  connectivity state and disposable generated preview data; cannot modify sources.
- **Preview result**: Compiler-produced web output linked to a source revision;
  retained last-valid output is distinct from current output.
- **Reference corpus**: One `smill` sheet first, then `bdock`/`whr` sheets with verified
  product identities, provenance, representative cards and optional cross-links.
- **Journey evidence**: Stage outcomes, command logs and failure-specific browser
  evidence; stored separately from deterministic site output and source documents.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: From a clean temporary workspace, one automated run completes all of
  Story 1's steps for the `smill` reference with no manual intervention. Adding the
  other two references uses the same harness and passes their shared checks.
- **SC-002**: For a documented reference workload of at most 100 sheets of at most
  100 KiB each, a saved edit becomes visible in the connected browser within one
  second in at least 95 of 100 measured updates on the documented reference machine.
  Initial preview becomes usable within three seconds in at least 19 of 20 starts.
- **SC-003**: In 100 rapid-save sequences, the final visible content and diagnostics
  match the latest saved revision in every case. Every intentionally invalid or
  disconnected state is visibly labeled; zero stale results are labeled current.
- **SC-004**: Every adversarial access/execution case in Story 5 and the edge-case
  corpus is refused or rendered inert as specified; selected source files remain
  byte-identical before and after each preview journey except external-editor saves.
- **SC-005**: Desktop and 390-pixel-wide reader journeys pass search, keyboard links,
  exact copy and appearance checks with zero horizontal page overflow. With browser
  scripting disabled, all expected reference titles, cards and navigation remain
  accessible. Two independent generated websites have identical file bytes.
- **SC-006**: At least four of five representative authors can start a session,
  locate a card, save an external edit and recover from a source error in under five
  minutes using the supplied instructions without maintainer assistance. This is
  a usability acceptance exercise, not a claim based on automated tests.

## Assumptions

- This invocation creates one feature specification: 007 local live preview, with
  its first delivery gate explicitly narrowed by the user's latest instruction to
  one `smill` sheet and one complete automated browser journey. Later stages in
  this feature add the two corpus inputs and then saved-file live preview; they
  must not delay or expand the first slice.
- Browser editing belongs to the later visual-editor feature. Initially authors
  edit in Helix or another external editor; Studio displays saved files only.
  Unsaved editor buffers are not observable, and no autosave or write endpoint exists.
- Existing source format, compiler semantics and Hugo behavior are dependencies;
  this feature does not redefine GFM+, add another editable format or change schema.
  Hugo is already supported. No dependency on the unimplemented PDF/navi milestones
  is introduced, and their implementation/deployment gates remain unchanged.
- Inputs follow existing nonrecursive file-selection boundaries. No project manifest,
  implicit `sheets/` convention, network fetching or general file browser is assumed.
- Automated test executables are temporary verification tools, not distribution
  builds. Site generation remains local; test evidence is not automatically uploaded.
- The `smill` reference is verified first. The intended BundleDock/WorkHarbor
  versions and their command inventories are pinned before adding those sheets;
  installed binaries or similarly named checkouts are not assumed authoritative.
- Performance workload and usability targets above are proposed acceptance targets,
  not existing measurements. Planning must document the measurement environment
  and any explicit resource ceilings without silently ignoring selected content.
- Specification preparation authorizes no implementation, deployment, PDF/navi
  generation or publication. Future roadmap and design proposals are recorded
  separately in `docs/studio-roadmap.md`; only 007 is specified by this invocation.
