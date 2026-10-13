# Plan — 012

Thread one normalized base path through `internal/render/hugo`: the generator,
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
Deployment stays out of scope and separately approved.
