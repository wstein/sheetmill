# SheetMill Studio roadmap and design inputs

Specification request: 2026-10-13. These are staged product intentions, not
implementation or deployment authorization. Updated 2026-10-14: the
[007 local live preview](../specs/007-live-preview/spec.md), three-sheet reader
corpus and automated performance checks are implemented; the five-author study
remains pending. Items 008–010 still have no specification directories.

## Delivery order

1. Prove one real `smill` sheet through the actual CLI, Hugo and an automated browser
   journey. Verify content, keyboard search/navigation, exact copied text, responsive
   layout, light/dark appearance and inert examples. No deployment or later outputs.
2. Add `bdock` and `whr` as content inputs to that same harness after pinning their
   intended versions and verifying command inventories; do not invoke their commands
   as part of compiling or browsing the references.
3. Complete 007 `smill serve`: local read-only saved-file preview, external-editor
   workflow, automatic refresh, diagnostic/freshness state and explicit file scope.
4. Proposed 008 Studio visual editor: source, split and structured card editing.
   Introduce explicit saves, dirty state, conflict detection and optional autosave
   only with a separately specified write model. Preserve the user's raw source,
   unknown metadata, card IDs and heading attributes; unsupported constructs remain
   editable as raw Markdown and never silently disappear on mode switches.
5. Proposed 009 `smill playground`: reuse Studio in an isolated disposable workspace,
   in memory by default, built-in examples, explicit local-file import and export,
   no project writes/autosave/Git operations. Import remains the controlled adapter.
6. Proposed 010 static online playground: optional public browser-local mode after
   production compiler parity is demonstrated. No accounts, hosted collaboration,
   remote compile service or public deployment is authorized by this roadmap.

The first slice must ship independently; server/editor work must not become a
prerequisite for testing the existing GFM → Hugo → browser workflow. Typst/PDF remains specification-only (005); navi text export (006) is implemented
separately, with sidecars authorized as a follow-up. Studio remains web-only.
PDF/navi preview integration requires its own scope and neither blocks 007 web preview.

## Architecture constraints to carry into planning

- Reuse the existing Go compiler and semantic IR across CLI, Studio and eventual
  output backends. Never introduce an approximate JavaScript parser/compiler as
  the authoritative preview for the editor.
- 007 uses the existing Hugo path to verify production web behavior. A later fast
  IR preview may reuse renderer code, but requires explicit parity tests and clear
  distinction from a publishing preview. It is not another source format.
- The user's preferred local-server design is native Go with a small bundled UI.
  File watching, server-sent events versus WebSockets, subprocess lifecycle and
  render cancellation are planning decisions, not selected dependencies here.
- Local sessions need loopback-only binding, strict origin/host validation, bounded
  resource use and containment throughout the session. The first milestone has no
  mutation endpoints. Before browser writes are introduced, specify unpredictable
  session credentials, request authorization, conflict handling and atomic saves.
  Never expose credentials in logs or test evidence. No terminal or snippet runner.
- Browser-only Go/WebAssembly compilation is a later feasibility question. A static
  public playground must demonstrate the same compiler contract; unsupported
  previews must be labeled unavailable rather than simulated. No shell execution,
  arbitrary URL fetching or remote host access may arise from source content.
- PDF previews would require the independently authorized Typst renderer and pinned
  installed tools; navi previews must remain inert text. Neither is part of 007.

## Editor proposals retained for later evaluation

The user proposed Milkdown for structured visual editing and CodeMirror for source
mode, with Vditor and TOAST UI Editor as alternatives. These are candidate design
inputs, not chosen packages, current compatibility claims or a dependency review.
Later research must verify maintenance, exact versions, licensing, accessibility,
bundle/offline behavior and round-trip fidelity against the SheetMill source corpus.

A dedicated card extension must preserve attributes such as `{.col-span-2}` and
explicit IDs. Switching source/visual modes without edits must preserve source
bytes; deliberate structural edits require defined minimal changes and loss-aware
handling of unsupported syntax. Unknown/imported metadata must not be dropped.
A free-form rich-text editor must not become the only editing mode.

## Naming and boundaries

SheetMill Studio names the graphical experience. `smill serve` opens existing
sources; `smill playground` later opens a disposable sandbox. Proposed flags such
as `playground --example helix` and `playground --import python.md` remain uncommitted
CLI contracts until their own specifications exist. Local previews and exports do
not imply publication, hosting, release builds or permission to redistribute imports.
