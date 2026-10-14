# Plan — 012

Delivered: the normalized base path and canonical origin/sitemap are already
implemented. The remaining design below concerns the general redistribution gate
and host-header guidance. Existing approved Pages publication is recorded in
AGENTS.md; this increment performs no deployment.

Original base-path approach: thread one normalized base path through `internal/render/hugo`: the generator,
route/link rewriting, search index destinations, brand/asset references in the
embedded templates, and `assets/app.js` (which currently fetches `/search.json`).
Prefer emitting the prefix via Hugo config/data so the theme does not hardcode it.
Hugo `baseURL` is derived from the option; the placeholder host stays only for the
path-only default.

Add a publication check in the CLI before any write: read the license metadata
already kept per sheet, compare with an explicit allow list file, and emit the new
diagnostic. The list is data, not code, and is itself reviewed. No network access.

Verification: failing tests first; build at `/` and a sub-path and scan the
output for root-relative URLs; serve locally (explicit local Hugo only) and
exercise navigation and search; CSP lint test over templates; repeated-hash
determinism; gofmt, go test, go vet, git diff --check; independent review.
New deployment stays out of scope and separately approved.
