# Feature 018: layout paths on GitHub Pages

**Status**: Specification revised after review rounds 2 and 3; the user confirmed the
new public paths on 2026-10-16, so the workflow change may be implemented on the topic
branch (see Authorization). Merge and deployment remain separately gated.
**Authorization** (the quote is not yet recorded in `AGENTS.md`; T001 adds it, until then it is unverified in the repo): the user's 2026-10-16 request ("publish the three layouts as
separate static paths, not a JS switcher") extends feature 012 (Pages) and uses the `--header-link` option (#39) and the
layouts of 004, 016. Two new public paths, `/sheetmill/demo-compact/` and
`/sheetmill/demo-cards/`, are a **new scope extension** of the 012 Pages approval
(constitution 7: broader publication stays gated). **The user confirmed that
extension on 2026-10-16** (answer "1j" to "confirm `/demo-compact/` and `/demo-cards/`
as new Pages paths"), so T005 (editing `pages.yml`) and T007 (docs that name the new
URLs) are unblocked for the topic branch; T001 records the confirmation with its date.
Together the two authorize, on a topic branch only, the specification, tests, verification script, workflow and docs edits, a topic push and a draft PR (the coordinator opens the PR). No merge, tag, release, deployment or workflow run is
authorized by this feature; merging to `main` is what would activate the changed
workflow and needs separate human approval.

## User scenarios

1. **Reader compares layouts.** A reader on the demo home opens another layout by an
   ordinary link. *Given* the editorial demo home, *when* the page is read, *then* it
   links to the compact and cards homes and not to itself; each variant home links to
   the other two (FR-2, rule 7).
2. **Existing links keep working.** *Given* a bookmark to `/sheetmill/demo/` or the
   manual's `Example library` link, *when* the new site is deployed, *then* both still
   reach the editorial demo (FR-1, rules 3 and 5).
3. **Maintainer is protected from a bad deploy.** *Given* a wrong base path, a layout
   marker mix-up or a page-set difference, *when* the workflow builds, *then* the job
   fails before upload and the previous deployment stays live (FR-6, FR-7, FR-9).

## Success criteria

- **SC-001** Three layouts exist at the three paths, each canonical to itself (FR-1, FR-4).
- **SC-002** Each build links exactly to the other two layouts (FR-2, rule 7).
- **SC-003** The verifier accepts the valid tree and the real manual-plus-demos tree
  and rejects a negative tree for each rule 1-8 (FR-7).
- **SC-004** Rebuilds are byte-identical (FR-8).
- **SC-005** No new external request, script, cookie or storage key (FR-11).
- **SC-006** The workflow structure is pinned by a parsing test (FR-1..3, FR-6, FR-7).
- **SC-007** Failure of any step leaves the previous deployment untouched (FR-9).

The FR-to-task table is in [analysis.md](analysis.md).

## Goal

The Pages site publishes the demo library once, in one layout. Publish the same
`examples/sheets` three times, one static site per layout, and let readers move
between them with ordinary links. No script, cookie or storage is added.

## Published paths

| Layout (`--theme`) | Base path | Site URL | Header-link labels shown |
| --- | --- | --- | --- |
| editorial (default) | `/sheetmill/demo/` | `https://wstein.github.io/sheetmill/demo/` | Compact layout, Cards layout |
| `classic-compact` | `/sheetmill/demo-compact/` | `https://wstein.github.io/sheetmill/demo-compact/` | Editorial layout, Cards layout |
| `reference-cards` | `/sheetmill/demo-cards/` | `https://wstein.github.io/sheetmill/demo-cards/` | Editorial layout, Compact layout |

The manual stays at `/sheetmill/` unchanged, including its `Example library` link to
`/sheetmill/demo/`. The default demo URL does not change, so existing links, the
README, llms.txt and the agent guide stay valid. Names are `demo-compact` and
`demo-cards` (siblings of `demo`, not `demo/compact/`) because a nested path would
require the editorial build, mounted at `/sheetmill/demo/`, to leave a hole for the
others and would make the default's relative layout depend on its variants.

## Requirements

- **FR-1 Builds.** `pages.yml` runs three `smill build hugo examples/sheets` exports
  from one source tree with `--base-path` and `--site-url` whose paths are equal
  (as 012 requires), the demo description and edit base of today, the title `SheetMill Demo` for the
  default and `SheetMill Demo (compact layout)` / `SheetMill Demo (cards layout)`
  for the variants (distinct `<title>` and `og:title` help readers and reduce
  clustering; no renderer change), and
  `--theme classic-compact` / `--theme reference-cards` for the variants. The default
  build passes no `--theme`, so it keeps the default layout by construction.
- **FR-2 Cross-links.** Each build passes exactly two `--header-link LABEL=URL`
  options naming the other two layouts, with the labels in the table and
  root-relative values that include `/sheetmill/` (the option does not prefix the
  base path). Links target the other layout's home page; per-page switching is out of
  scope. Two links leave room under the four-link limit for later additions. Every
  label is unique within a build and the link set never contains the build's own
  layout.
- **FR-3 Same inputs, same clock.** All Hugo invocations use the identical pinned
  binary, `--clock 2000-01-01T00:00:00Z` and `--noBuildLock --quiet`. The three
  exports differ only in `--theme`, `--base-path`, `--site-url`, `--out`, `--title`
  and `--header-link` (checked by the workflow test); the page set is the same
  (rule 6).
- **FR-4 Canonical and duplicate content.** Each variant is canonical for itself:
  `--site-url` equals its own base URL, so `rel=canonical`, `og:url`, `og:image` and
  its `sitemap.xml` use `https://wstein.github.io/sheetmill/demo-<x>/`. Variants
  carry no `noindex`. Reasons: the layouts differ in markup and navigation (they are
  distinct presentations, not mirrors); `noindex` would need a new renderer option
  and `noindex` combined with a canonical to another URL is discouraged; a canonical
  pointing at the default would also need a new option (deferred, see Open decisions).
  Search engines may still cluster near-duplicates; that is accepted for a demo.
- **FR-5 Sitemap and robots.** Each export keeps emitting its own `sitemap.xml` that
  lists only URLs under its base path, as the demo does today (the manual's own sitemap is checked in rule 4). The manual's
  `/sheetmill/sitemap.xml` continues to list only the manual: this is a deliberate
  choice to avoid a renderer change, since such a sitemap could technically list
  URLs below its path. No sitemap index and no `robots.txt` are added: a
  project site cannot control the host-root `robots.txt`, and 012 FR-007 keeps robots
  disabled. The variant sitemaps are therefore not advertised; discovery is by the
  header links; an owner may submit the variant sitemaps by hand. This is an explicit exclusion, not an omission.
- **FR-6 Mount and collisions.** The variants are mounted below the manual output.
  The workflow refuses to build if `public/demo`, `public/demo-compact` or
  `public/demo-cards` already exists after the manual build (a manual page of that
  slug), and the mount order is manual first, then the three demos. The prefix `demo`
  is reserved at the top level: no manual slug may start with `demo` (for example
  `demonstration`), because rules 1 and 3 test the prefix; the guard is the exact
  names above, the verifier enforces the whole prefix.
- **FR-7 Post-build verification.** A separate workflow step named `Verify the Pages
  tree` (not the tail of the build run block) is the last step before
  `upload-pages-artifact`, after the build step (which ends with the `llms.txt` and agent-guide copies).
  It runs `sh .github/scripts/verify-pages.sh "$RUNNER_TEMP/public"`, which fails the
  job on any violation. The script reads the generator's double-quoted attributes. A
  test in `internal/render/hugo` (always run, no Hugo needed) scans the theme
  templates and fails on a single-quoted or unquoted `href`, `src` or `action`
  attribute. The script also carries a tag-anchored check that fails on
  `<[a-zA-Z][^>]* (href|src|action)=[^"]`, which catches single-quoted or unquoted
  attributes in the generated HTML, including the markup that `hugo.go` writes from Go
  format strings (for example the `<a href=` strings near lines 365, 385 and 768),
  which the template scan cannot see; escaped prose uses `&lt;`, so it has no false
  positives, and it matches nothing in the current sources. The real-manual run (T006)
  exercises the generated HTML. The script
  uses `sh` with `grep -E`, `find`, `sort` and `tr` only, reads the local
  tree, writes nothing (no temporary files) and makes no network request; tool errors
  fail the run and the root may contain only `[A-Za-z0-9/._-]` (else exit 2). Why a shell script and not Go: it must run on the
  built output tree in CI with no extra build step, its checks are plain text
  matching that reviewers can read in one screen, and it adds no public CLI surface or
  package. The Go side is limited to the test-only package `internal/pagesworkflow`
  (constitution 5: standard tools first, a small justified exception, no new
  dependency; `mvdan.cc/sh/v3` is already in `go.mod`). A page belongs to the demo whose directory (`demo`,
  `demo-compact`, `demo-cards`) contains it, else to the manual. Rules:
  1. the top level of `public/` holds the manual output plus `demo`, `demo-compact`,
     `demo-cards`, `llms.txt` and `agent-guide.md`. The real check is a subset check:
     the three demo directories, `llms.txt`, `agent-guide.md`, `index.html`,
     `sitemap.xml`, `search.json`, `assets` and `brand` exist, and no other entry
     starts with the reserved prefix `demo` (FR-6); other manual entries are not
     enumerated;
  2. each of the manual and three demos has `index.html`, `sitemap.xml`,
     `search.json` and `assets/`;
  3. every root-relative `href`/`src`/`action` value in HTML begins with its own base
     path including the trailing slash (`/sheetmill/demo/` is not a prefix of
     `/sheetmill/demo-compact/`), or equals one of the three layout homes exactly.
     For the manual (`/sheetmill/`), any value beginning `/sheetmill/demo` must equal
     a layout home exactly or lie under one of the three homes;
  4. every absolute `canonical`, `og:url`, `og:image` and sitemap `<loc>` of a demo
     begins with that demo's own site URL including the trailing slash, every manual sitemap `<loc>` begins with `https://wstein.github.io/sheetmill/` and none with `https://wstein.github.io/sheetmill/demo`, and the
     string `example.invalid` (the Hugo `baseURL` placeholder in `hugo.go`) appears
     nowhere in the output tree (a leak would mean a canonical was not rewritten);
  5. layout markers on every HTML page: the cards demo has the body class tokens
     `classic` and `reference-cards` and stylesheets matching
     `assets/classic\.[0-9a-f]{64}\.css` and `assets/cards\.[0-9a-f]{64}\.css`;
     the compact demo has the token `classic` and the classic stylesheet only;
     the editorial demo and the manual have no `classic` token and neither stylesheet;
  6. the three demos have the same set of `*.html` paths (assets legitimately
     differ by theme);
  7. the manual home still links `/sheetmill/demo/`; the header-link elements of each demo home (`nav.header-links a`,
     `a.topbar-link` other than `Library`, `p.hero-links a`) point exactly at the
     other two layout homes exactly once each (a repeat fails), never at the demo's own home (the logo and `Library`
     link to the own home legitimately and are not header links);
  8. no `href`, `src`, `srcset` or `action` value is protocol-relative (starts with `//`);
     every absolute `src`, `srcset`, `action` or `<link href>` value (any scheme,
     case-insensitive) begins with `https://wstein.github.io/sheetmill/`; no `<base>`
     element exists. Plain `<a href>`
     links inside prose are exempt because documents may show loopback examples such
     as `http://127.0.0.1:1313/` (the autolink in `docs/hugo.md`); T006 runs the script on
     the real manual built from `docs/manual-pages.txt`, so this case is a regression
     test, not a stand-in.
- **FR-8 Determinism.** The same repository revision produces byte-identical output
  for each of the three demo trees (existing 012 / 004 FR-005 guarantee, now shown
  for three builds). The verification script is deterministic and writes nothing.
- **FR-9 Failure and rollback.** Any export, Hugo, collision or verification failure
  fails the `build` job; `deploy` has `needs: build`, so the previous deployment
  stays live untouched. Rollback is reverting the change on `main` and letting the
  main-only workflow redeploy, which a human approves like any merge. Removing a
  layout later makes its path return 404; no redirect is promised.
- **FR-10 Cost.** Measured locally with Hugo 0.167.0: demo 82 files / 0.84 MB,
  compact 83 files / 0.88 MB, cards 84 files / 0.89 MB; gzip tar about 0.19 MB each;
  whole site 3.0 MB (was about 1.2 MB). Three extra Hugo builds take seconds; the
  10-minute job limit and Pages size limits are far from reached.
- **FR-11 Security.** No new external request, script, cookie, storage key or remote
  font. No inline script/style is added, so a strict CSP stays valid. Header-link
  values are validated by the existing option; the Pages artifact still contains only
  the generated sites plus `llms.txt` and `agent-guide.md`. Workflow permissions,
  action pins, triggers and the `github-pages` environment are unchanged.
- **FR-12 Authorization records.** Constitution 7: two new public paths are a scope
  extension, not part of the already authorized demo. `AGENTS.md` (the 012 bullet) and
  specs/012 (status line, FR-009, Resolved decisions) record
  `/sheetmill/demo-compact/` and `/sheetmill/demo-cards/` as an **extension of the 012
  approval, confirmed by the user on 2026-10-16** (the confirmation covers the paths,
  not their activation). The paths go live only through a separately approved merge to
  `main`; no line may say that merge, a deployment or a workflow run is approved.
- **FR-13 Documentation.** `docs/hugo.md` (Pages section and Search engines
  paragraph), `llms.txt`, `docs/agent-guide.md` and the README list the layout URLs.
  `docs/cli.md` needs no change because no flag changes.

## Known limitations (accepted)

- Editorial pages carry the cross-layout links in `nav aria-label="Related sites"`.
  In the classic layouts they are `.topbar-link` anchors, which `classic.css` hides
  below 640px; on mobile the switch exists only through `p.hero-links` on the home
  page, not on sheet pages. Accepted for the demo.
- No page names its own layout; the title suffix (`compact layout`, `cards layout`)
  and the layout look are the only cue. Naming the current layout on each home page
  needs a renderer change and is deferred.

## Out of scope

A JavaScript or cookie layout switcher, per-page layout links, remembered layout,
redirects, a canonical-to-default option, `noindex`, a sitemap index, robots output,
new renderer options, other sheets, a custom domain, and any deployment, tag or
release. No change to `internal/` renderer code is planned.

## Testing

- A Go test parses `pages.yml` (YAML, not regex over the whole file) and asserts the
  three export commands, base-path/site-url agreement, theme flags, header-link
  rules, shared clock, collision guards, the verification step order (after builds,
  before upload) and that triggers/permissions are unchanged.
- Table tests in `internal/pagesworkflow` run `sh .github/scripts/verify-pages.sh`
  through `os/exec` against synthetic trees: one valid, and one failing tree per rule
  1-8 (including the `127.0.0.1:1313` prose link that must pass and a `//` src, an
  `http:` src, a single-quoted or unquoted `href`, `example.invalid` and a `demonstration` entry that must fail). If `sh`
  is not available the tests are skipped with a message, and a skip is not a pass.
- With `SHEETMILL_HUGO` set, an integration test builds the real manual from
  `docs/manual-pages.txt` and the three real demos through `internal/cli.Run`, mounts
  them as the workflow does, runs the script, and compares two builds.
- An always-run test in `internal/render/hugo` checks that the theme templates quote
  every `href`, `src` and `action` attribute (FR-7).
- Browser checks are not part of this feature; Node is not required.

## Open decisions (recommendations)

1. Canonical of variants to the default demo needs a new `--canonical-site-url`
   option: defer; revisit only if search analytics show duplicate-content harm.
2. Add a `Manual` back link to each demo (three links): recommended later, not done
   here to keep the change to the request.
3. Page-level switching (same sheet in another layout) would need a renderer option
   for a path-preserving link: defer.
