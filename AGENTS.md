# SheetMill agent policy

SheetMill is a standalone Go reference compiler. The CLI is `smill`.
Use exact product spelling in prose: SheetMill, BundleDock, WorkHarbor, CrewBook.
Sibling repository policies do not govern this repository.

## Scope and authorization

- Routine local source/document edits, dependency inspection, formatting, tests,
  fuzzing, vet and source/security checks are allowed within the requested work.
- The current milestone is the parser, validator, deterministic JSON IR and CLI.
  Do not implement web, PDF, Typst or navi renderers yet.
- Use `codex/` feature branches and Conventional Commits. Topic pushes and draft
  PRs belong to an authorized implementation handoff; otherwise ask before publishing.
- Merge, tag, release, deploy, binary/distribution builds, website/PDF generation,
  artifact publication and sibling-repository edits require separate human approval.
- Do not execute source snippets, fetch source-linked resources, weaken branch
  protections, add automatic merges or publish artifacts from CI.
- Unknown source metadata must be preserved or diagnosed, never silently dropped.

## Implementation and verification

Use Go's standard library first; justify and pin dependencies. Keep diagnostic
codes and JSON schema changes deliberate and documented. Use source-positioned
errors, table-driven negative tests, independently authored fixtures, fuzz targets
and byte-for-byte golden JSON. Write failing behavior tests before implementation.

Run `gofmt`, `go test ./...`, `go vet ./...` and `git diff --check` before review.
Golden changes require semantic inspection; regenerate explicitly with
`go test ./internal/markdown -run TestGolden -update`.
Report commands actually run and remaining gaps. Independent human or separately
assigned reviewer approval is expected before merge. An author's own checks do
not constitute independent review. Delegation is not required by this policy.

## Licensing

Original code uses EUPL-1.2 SPDX headers. Keep the full license and provenance.
Implement compatibility independently; do not copy Reference's GPL implementation,
styles or authored sheets. No final branding palette has been selected.
