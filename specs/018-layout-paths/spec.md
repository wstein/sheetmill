# Feature 018: layout paths on GitHub Pages

**Status**: Specification; implementation follows independent review.
**Authorization**: user, 2026-10-16 ("publish the three layouts as separate static
paths, not a JS switcher"; topic push and draft PR). Extends feature 012 (Pages) and
uses the `--header-link` option (#39) and the layouts of 004, 016. No merge, tag,
release, deployment or workflow run is authorized by this feature. Merging to `main`
is what would activate the changed workflow; that merge needs separate human approval.

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
  (as 012 requires), the demo title/description and edit base of today, and
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
  binary, `--clock 2000-01-01T00:00:00Z` and `--noBuildLock --quiet`. The variants
  differ from the default only by theme assets/markup and by base path, header links
  and canonical URLs.
- **FR-4 Canonical and duplicate content.** Each variant is canonical for itself:
  `--site-url` equals its own base URL, so `rel=canonical`, `og:url`, `og:image` and
  its `sitemap.xml` use `https://wstein.github.io/sheetmill/demo-<x>/`. Variants
  carry no `noindex`. Reasons: the layouts differ in markup and navigation (they are
  distinct presentations, not mirrors); `noindex` would need a new renderer option
  and `noindex` combined with a canonical to another URL is discouraged; a canonical
  pointing at the default would also need a new option (deferred, see Open decisions).
  Search engines may still cluster near-duplicates; that is accepted for a demo.
- **FR-5 Sitemap and robots.** Each export keeps emitting its own `sitemap.xml` that
  lists only URLs under its base path, as the demo does today. The manual's
  `/sheetmill/sitemap.xml` continues to list only the manual (a sitemap may not list
  URLs outside its own path). No sitemap index and no `robots.txt` are added: a
  project site cannot control the host-root `robots.txt`, and 012 FR-7 keeps robots
  disabled. The variant sitemaps are therefore not advertised; discovery is by the
  header links. This is an explicit exclusion, not an omission.
- **FR-6 Mount and collisions.** The variants are mounted below the manual output.
  The workflow refuses to build if `public/demo`, `public/demo-compact` or
  `public/demo-cards` already exists after the manual build (a manual page of that
  slug), and the mount order is manual first, then the three demos.
- **FR-7 Post-build verification.** Before upload, `.github/scripts/verify-pages.sh`
  checks the assembled `public/` tree and fails the job on any violation:
  1. each of the three sites has `index.html`, `sitemap.xml`, `search.json` and an
     `assets/` directory;
  2. every root-relative `href`/`src`/`action` in HTML starts with the page's own
     base path, or equals the home path of one of the three layouts (header links);
  3. every absolute `canonical`, `og:url`, `og:image` and sitemap `<loc>` of a demo
     starts with that demo's own site URL, and none names another demo's path;
  4. the layout markers match: `reference-cards` body class and `cards.*.css` only in
     the cards demo; `classic.*.css` in the compact and cards demos only;
  5. the three demos have the same set of page paths (same sheets, same tags);
  6. each demo home links to the other two layout homes and not to itself;
  7. no `http://` resource attribute and no `//host` URL appears in `href`, `src` or
     `action` attributes.
  The script reads only the local tree; it performs no network request.
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
- **FR-12 Documentation.** `docs/hugo.md` (Pages section and Search engines
  paragraph), `llms.txt`, `docs/agent-guide.md` and the README list the layout URLs.
  `docs/cli.md` needs no change because no flag changes.

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
- Table tests run `verify-pages.sh` against synthetic trees: one valid, and one
  failing tree per rule above.
- With `SHEETMILL_HUGO` set, an integration test builds the three real demos, mounts
  them under a manual stand-in and runs the script, and compares two builds.
- Browser checks are not part of this feature; Node is not required.

## Open decisions (recommendations)

1. Canonical of variants to the default demo needs a new `--canonical-site-url`
   option: defer; revisit only if search analytics show duplicate-content harm.
2. Add a `Manual` back link to each demo (three links): recommended later, not done
   here to keep the change to the request.
3. Page-level switching (same sheet in another layout) would need a renderer option
   for a path-preserving link: defer.
