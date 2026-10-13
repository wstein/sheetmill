# 012 — Hosted reference library: base path and publication gate

Created: 2026-10-13 · Status: draft, specification only
Depends on: 004-hugo-renderer (implemented); content rights of the chosen sheets

## User scenarios and acceptance

1. P1: An author builds the Hugo project for a chosen base path (for example
   `/sheets/`) and serves it from that sub-path. Navigation, assets, brand files,
   search and permalinks all work; nothing still points at the origin root.
2. P1: An author runs a publishable build. Any sheet whose redistribution rights are
   unknown fails with an actionable diagnostic unless explicitly allow-listed.
3. P1: A maintainer reviews a publication candidate. The site is byte-identical for
   identical input and configuration, and carries no tracking or remote resources.
4. P2: A host operator applies a strict Content-Security-Policy without breaking
   the site, because pages use no inline scripts or styles.

## Requirements

- FR-001: Add a validated base path / base URL option to the Hugo export. Reject
  query strings, fragments, `..` segments, backslashes, control characters, userinfo
  and absolute host paths. Normalize to a leading and trailing `/`. The default stays
  origin root so current output is unchanged.
- FR-002: Every generated link, stylesheet, script, favicon, brand image, search
  fetch (`/search.json`, issued by `assets/app.js`) and permalink respects the base
  path. The generated config `baseURL` (today `https://example.invalid/`) is derived
  from the option; an absolute origin is optional and, when given, must be https.
- FR-003: No host paths, usernames or clock values leak into output. Output stays
  deterministic and independent of input order (as in 004 FR-005).
- FR-004: Per-sheet redistribution review. Import provenance records licenses as
  supplied and states parsing does not approve redistribution. A publishable build
  requires, for each sheet, either a recorded license in an explicit allow list
  or an explicit maintainer-signed exception; otherwise it fails.
- FR-005: Introduce a deliberate diagnostic, tentatively `REN004` (error: sheet
  not cleared for redistribution), and a warning for missing or non-allow-listed
  license metadata outside publishable mode. Document in docs/cli.md and
  docs/hugo.md; changing the code set follows the repository diagnostics policy.
- FR-006: Pages must remain CSP-friendly: no inline scripts, inline event handlers
  or `style=` attributes. Add a test that fails on any such addition. (Current
  templates were found to use only external `/assets/app.js` and `site.css`.)
- FR-007: Decide and document robots and sitemap. Both are disabled today
  (`enableRobotsTXT: false`, `sitemap` kind disabled); enabling either requires the
  absolute origin from FR-002 and a documented rationale.
- FR-008: No analytics, tracking, cookies or remote fonts/CDNs (004 FR-008 stays).
- FR-009: Publication and deployment remain separately human-approved per
  AGENTS.md. `smill` never uploads. CI never publishes artifacts or merges
  automatically. This feature only produces a reviewable local build.
- FR-010: Document security-header guidance (CSP, `X-Content-Type-Options`,
  `Referrer-Policy`) for host operators as recommendations, not enforced behavior.

## Success criteria

Tests build the same sheets at root and at a sub-path and assert no root-relative
URL remains in the sub-path output; a served sub-path check exercises navigation and
search. Negative table tests cover invalid base paths and unlicensed sheets.
Repeated builds hash identically. An independent reviewer approves the exact final
revision; no deployment or merge happens without human approval.

## Open decisions

- Hosting target (static host, project pages, self-hosted) and who operates it.
- Intended base path and whether a custom domain is used.
- Which sheets are licensed for redistribution, and the allow-list format.
- Whether robots/sitemap are enabled once an origin is known.
- Whether the diagnostic is an error or warning by default.

Relationship: a proposed future static online playground (roadmap item 010) would
depend on this feature's base path and publication gate; it is out of scope here.
