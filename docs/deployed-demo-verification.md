# Deployed demo verification

Verified on 2026-10-13 after [PR #9](https://github.com/wstein/sheetmill/pull/9)
merged into `main` at `b13df7a`. The Pages [deployment run](https://github.com/wstein/sheetmill/actions/runs/38447095867)
succeeded. The inspected public URL was <https://wstein.github.io/sheetmill/>.
This was read-only verification of the existing deployment, not a new deployment.

Chromium 156.0.8078.4 / Playwright 1.64.0 / Node 24.21.0 exercised desktop
1280×800, mobile 390×844 and JavaScript-disabled desktop contexts. All passed:

- Homepage and all three library links; SheetMill CLI page and stable card link.
- Search followed by Tab/Tab/Enter keyboard navigation to the intended result.
  The query `Build a local library` returned two matches; Tab reached the first
  and a second Tab reached SheetMill CLI. The initial one-result expectation was
  corrected after inspecting the real three-sheet library.
- Exact displayed code and actual clipboard readback after seeding a sentinel.
- Applied light/dark background colors and preference persistence after reload.
- Visible EUPL provenance and no horizontal page overflow.
- Ordinary navigation/content with JavaScript disabled.
- No page errors, HTTP error responses or requests outside `/sheetmill/`.

Search, clipboard and appearance assertions apply to both scripted viewports;
the no-JavaScript context checks basic content/navigation only. Commands were
copied as text and never executed. Local script/results are preserved under
`.local/verify-demo.mjs` and `.local/deployed-demo-verification.json` in the corpus
worktree. This is a point-in-time Chromium check, not Safari/Firefox coverage or
an availability monitor.
