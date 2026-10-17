# Feature 019: tag-driven release workflow

**Status**: Specified and implemented on the topic branch `codex/019-release-workflow`; independent
review rounds are in [analysis.md](analysis.md). Nothing has been run on GitHub.
**Authorization** (recorded in `AGENTS.md`, T001): the owner, Werner Stein, decided in
chat on 2026-10-17: "017/018 in v0.1.0, ... tag v0.1.0 => CI based release workflow,
update CI pipeline". This feature reads that as three decisions:

1. v0.1.0 contains features 017 and 018, so the existing annotated tag `v0.1.0`
   (-> `05eba4d`, created by the owner on 2026-10-16, no GitHub release) is deleted and
   re-created by the owner on a later `main` commit (see [Retagging](#retagging)).
2. Pushing a version tag triggers a CI workflow that builds, verifies and publishes the
   release.
3. The CI pipeline is updated to match.

The decision replaces the `AGENTS.md` rule "do not publish artifacts from CI" **for this
workflow only**. An owner-pushed tag matching `v[0-9]+.[0-9]+.[0-9]+` is the publication
approval for that tag's release assets; the workflow pattern is general, but whether the quote also approves every future tag is an open owner confirmation
(open decision 6), recorded in `AGENTS.md` as pending. It authorizes, on the topic branch, the spec,
tests, scripts, `release.yml`, the `check.yml` comment, docs, a topic push and a
coordinator-opened draft PR. It does **not** authorize: merging the workflow PR,
creating, moving or deleting any tag, running the workflow (a dry run is started only by the owner through `workflow_dispatch`;
the workflow has no `pull_request` trigger, so the draft PR starts no run),
creating a release, adding secrets, variables or environment protection rules, or
publishing to any channel except GitHub Releases of `wstein/sheetmill`. Constitution 7
text ("a spec or automated check grants none") is respected: the grant is the owner's
quote, recorded in `AGENTS.md`, not this spec. Amending constitution 7 to name the
release workflow beside the Pages workflow needs a human-approved change and is left to
the owner (open decision 1).

## User scenarios

1. **Owner releases.** *Given* the workflow is on `main`, *when* the owner pushes the
   annotated tag `v0.1.0` on a green `main` commit, *then* the workflow verifies the tag,
   builds three binaries, smoke-tests each natively, and creates a GitHub release with three archives and `SHA256SUMS` (FR-1..FR-6).
2. **Wrong tag is refused.** *Given* a lightweight tag, a tag on a commit that is not on
   `main`, a tag whose commit has no green `check` run, a tag pushed by someone other than
   the repository owner, or a tag whose release already exists, *then* the verify job fails
   and nothing is built or published (FR-2).
3. **Maintainer exercises the pipeline safely.** *Given* the owner starts the workflow by hand
   (`workflow_dispatch`), *then* the same gates, builds and smoke tests run as a dry run that
   cannot publish (FR-1, FR-7). No pull request or branch push starts it.
4. **Downloader verifies.** *Given* a release, *when* a user downloads an archive and
   `SHA256SUMS`, *then* `sha256sum --check` passes and `smill version` prints
   `smill v0.1.0` (FR-5).

## Success criteria

- **SC-001** Only an exact `vMAJOR.MINOR.PATCH` tag push can reach `publish` (FR-1, FR-6).
- **SC-002** Every verify condition has a script or workflow test that fails when the
  condition is removed (FR-2, FR-8).
- **SC-003** The release contains exactly the three archives and `SHA256SUMS`; checksums
  are computed once at build time and re-checked at smoke and publish (FR-5).
- **SC-004** Each shipped archive was extracted and run natively (`smill version` equals
  the tag, `--help`, `check`, deterministic `build hugo` export) before publish (FR-4).
- **SC-005** Only the `publish` job holds `contents: write`; no secret is used (FR-7).
- **SC-006** `go test ./...` fails if the workflow is weakened as listed in FR-8.

The FR-to-task table is in [analysis.md](analysis.md).

## Goal

Make a release a reviewable, repeatable CI run instead of a local build: the owner pushes
a tag, the workflow proves the tagged commit is releasable, produces the binaries from
that commit only, and publishes them with checksums. Nothing is signed or notarized.

## Requirements

- **FR-1 Triggers.** `release.yml` runs on:
  - `push` of tags matching exactly `v[0-9]+.[0-9]+.[0-9]+` (GitHub glob; the whole ref
    name must match, so `v1.2.3-rc1` triggers nothing; pre-releases are out of scope);
  - `workflow_dispatch` with one string input `version` (default `v0.0.0-dryrun`).
  A dispatch is a **dry run**: it runs `verify` (tag-specific checks skipped), `build` and
  `smoke`, and `publish` is skipped. Dry-run workflow artifacts live one day and, in a public repository, can be downloaded by any signed-in GitHub user. No branch
  `push` and no `pull_request` trigger (a PR trigger would build binaries on the draft PR
  without the owner's approval of a distribution build).
- **FR-2 Verify job** (`contents: read`, `actions: read`). In every run kind first: the version is
  resolved and `github.triggering_actor` must equal `github.repository_owner` (so a dispatch
  is owner-only too) and `github.actor` (the original starter of a re-run) must too, so an owner re-run of someone else's run does not count as approval. On a tag push, then in order:
  (a) the tag name matches the strict pattern `^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`;
  (b) the tag object is an annotated tag (`git cat-file -t refs/tags/<tag>` is `tag`), after
  fetching the tag explicitly;
  (c) the tagged commit equals `GITHUB_SHA` and is an ancestor of `origin/main` (full
  history, `fetch-depth: 0`);
  (d) (the actor check above, kept as a separate earlier step);
  (e) a run of the workflow file `check.yml` (queried by file, not by display name) for that
  commit with `event=push` has conclusion `success` (`gh api`, read-only); a run in
  progress, cancelled or failed fails the job with an instruction to wait for or re-run `check`;
  (f) no published release exists for the tag (`gh api repos/.../releases/tags/<tag>` returns
  404; any other error fails closed). Drafts are invisible to this read-only token and are
  caught by the publish job's own check (FR-6).
  In every run (dry or not): `.go-version` equals the `go` line of `go.mod`; `make check`,
  `make mod-verify` and `make test-assets` pass. The version is the tag name; for a dry run
  it is the `version` input or the default, checked against
  `^v[0-9]+\.[0-9]+\.[0-9]+-[0-9A-Za-z.]+$` (a suffix is mandatory, so a dry run can never
  carry a release version; the packaging script accepts both forms). Inputs reach shell only through `env:`.
  The job outputs `version`, `dry_run` and `commit`.
- **FR-3 Build job** (`needs: verify`, `contents: read`). Matrix exactly the `check.yml`
  `build` targets: `linux/amd64`, `linux/arm64`, `darwin/arm64`. On `ubuntu-24.04`, with
  `CGO_ENABLED=0`, `GOTOOLCHAIN=local`, `GOFLAGS=-mod=readonly`, `go build -trimpath -ldflags
  "-X github.com/wstein/sheetmill/internal/cli.Version=<version>" ./cmd/smill`. Each target
  is built twice, the second time with an empty `GOCACHE`, and the two binaries must be
  byte-identical (`cmp`). The only network access is the Go module proxy, whose content
  `go.sum` and `go mod verify` check, plus the Go toolchain `actions/setup-go` installs from
  the pinned `.go-version` (`GOTOOLCHAIN=local` forbids any other toolchain download). No other tool is downloaded. darwin/amd64 (end of life) and
  Windows are not built.
- **FR-4 Smoke job** (`needs: [verify, build]`, `contents: read`). One job per target on a
  native runner: `ubuntu-24.04` (linux/amd64), `ubuntu-24.04-arm` (linux/arm64), `macos-15`
  (darwin/arm64, Apple silicon). It downloads the archive, checks its checksum, extracts
  it and runs `.github/scripts/release-smoke.sh`: `smill version` equals `smill <version>`;
  `smill --help` exits 0; `smill check examples/sheets` exits 0; `smill build hugo
  examples/sheets` run twice yields identical trees with more than zero files. Hugo is not
  installed or run. The archive contains the files promised in FR-5.
- **FR-5 Archives.** Per target `smill_<version>_<os>_<arch>.tar.gz` where `<version>` is the
  tag without the leading `v` (for example `smill_0.1.0_linux_amd64.tar.gz`). It contains one
  top directory of the same name with `smill` (mode 0755), `LICENSE` and `README.md`
  (mode 0644). Packaging is deterministic: sorted names, owner 0, numeric owner,
  `mtime` = commit time, `gzip -n`. A sidecar `<archive>.sha256` is written at build time.
  The publish job concatenates the sidecars, sorted, into `SHA256SUMS` and runs
  `sha256sum --check` on it. Release assets are exactly the three archives and `SHA256SUMS`.
- **FR-6 Publish job** (`needs: [verify, build, smoke]`; runs only when the event is a tag
  `push` and `dry_run` is `false`; `permissions: contents: write`; `environment: release`).
  Steps: checkout (full history, no persisted credentials); download the artifacts; require
  `github.triggering_actor` to be the repository owner again (a re-run by someone else must
  not reuse the owner's earlier approval); list all releases (drafts included) and fail on any
  with the tag name;
  confirm the remote tag still peels to the verified commit (`git ls-remote`); build
  `SHA256SUMS`; generate `notes.md` with `.github/scripts/release-notes.sh`; run
  `gh release create <tag> --verify-tag --title "SheetMill <tag>" --notes-file notes.md
  [--draft] <assets>`. The release is **not** a draft by default. Setting the repository
  variable `SHEETMILL_RELEASE_DRAFT` to `true` adds `--draft` (the owner then publishes
  by hand). `--latest` is left to GitHub's default. The notes list version, commit,
  platforms, Go version, the checksum command, the statement that nothing is signed or
  notarized, and the commit subjects since the previous `v*` tag (or "First release.").
- **FR-7 Security posture.** Top-level `permissions: contents: read`; no job other than
  `publish` has `contents: write`; no job has `id-token`, `attestations`, `packages`,
  `actions: write` or `pull-requests`. Every `uses:` is pinned to a full commit SHA with a
  version comment (reusing `check.yml`/`pages.yml` pins; new pins in [research.md](research.md)).
  Every checkout sets `persist-credentials: false`. `concurrency` group is per ref with
  `cancel-in-progress: false`. Every job has `timeout-minutes`. No `continue-on-error`, no
  `if: false`, no `secrets.*`; `GH_TOKEN` is `github.token` only. `upload-artifact` is used
  only in `build`, `download-artifact` only in `smoke` and `publish`. `check.yml` stays
  `contents: read`.
- **FR-8 Tests.** A test-only Go package `internal/releaseworkflow` parses `release.yml`
  (YAML; `run:` blocks are checked by exact line and regular-expression rules, not a shell
  parser) and runs the scripts in temporary git repositories. A single `violations(workflow)` function returns
  every broken rule; mutation tests apply one weakening each to the real file text and
  require the matching violation: trigger widened or narrowed, pre-release pattern added,
  `contents: write` moved to another job, a tag-mutable action ref, missing
  `persist-credentials`, `continue-on-error` or `if: false` on a gate, matrix gaining or losing
  a platform, `publish` losing a `needs` entry or the `release` environment, `ldflags`
  removed, an extra `upload-artifact`, `--verify-tag` removed, draft default changed, a
  `secrets.` reference, each verify condition (a)-(f) removed or left unguarded, a `${{`
  expression inside a `run:` script, `publish` made unconditional (`always()`, no event or ref
  test), the publish actor, draft-duplicate, tag-race (`ls-remote`) and `sha256sum --check`
  steps removed, the `cmp` reproducibility check removed, a missing `timeout-minutes`, and
  `cancel-in-progress: true`. `cmd/smill` keeps its `check.yml` build-job tests; a new test
  asserts the `check.yml` and `release.yml` platform sets are equal.
- **FR-9 Pipeline update.** `check.yml` keeps its job names (they may be required checks),
  permissions, the compile-only `build` matrix and the `v0.0.0-ci` stamp. It gains a header
  comment naming `release.yml`. The parity test is the coupling. The dry run lives in
  `release.yml` (research, decision D4).
- **FR-10 Documentation.** `docs/releasing.md` (owner procedure), `README.md` (replace
  "no published binaries"), `docs/getting-started.md` (install from a release), `docs/cli.md`
  (release build note), `specs/README.md` row, `llms.txt` (release fact), and the
  `AGENTS.md` line (T001).

## Retagging

Owner-only; agents never touch tags. The old tag is already known to the Go module proxy
(`proxy.golang.org` serves `v0.1.0` for `05eba4d`), see research D9 and open decision 2.
Procedure (details in `docs/releasing.md`): merge the workflow PR; pick the `main` commit
(it must contain 017, 018 and the workflow, and have a green `check` run); delete the tag
locally and on origin; create an annotated tag on that commit; push it; approve the
verify. The `release` environment has no reviewer rule (owner decision 2026-10-17).

## Known limitations (accepted)

- No reviewer gate: the `release` environment is unprotected, so an owner tag push publishes
  as soon as verify, build and smoke pass.

- The workflow tests are string- and line-level rules plus mutations, not a shell AST. A
  determined edit can still disarm a check; they catch accidental or casual weakening, and
  the independent review plus the owner's review of the workflow PR remain the real gate.
  `testdata/release.golden.json` pins the whole parsed workflow as a **change detector**: it
  can be regenerated (`UPDATE_RELEASE_GOLDEN=1`, refused in CI, the run then fails on purpose),
  so it proves nothing about safety; the semantic rules do. Mutation rows match raw text, so
  action bumps and reformatting need test updates (docs/releasing.md).
- Every owner check compares with `github.repository_owner`; if the repository becomes
  organization-owned or is transferred, the checks fail closed until they are rewritten.

- No signing, notarization or attestation. On macOS a browser download may be quarantined
  by Gatekeeper; `curl`/`tar` downloads are not.
- Reproducibility is verified on one runner with one toolchain (two builds, one fresh
  cache); independent-machine reproduction is not claimed.
- The verify job re-runs `make check`, not the Hugo integration, race or fuzz tests; it
  relies on the green `check` run for the same commit for those.
- Tag-path behavior (annotated-tag fetch, API calls, environment gate, `gh release create`)
  cannot be exercised before the first real tag; scripts are tested locally and the dry run
  covers build and smoke only.
- A `v*` tag deleted and re-created does not refresh the Go module proxy.

## Out of scope

Windows, darwin/amd64, linux/386 or musl-specific builds, Homebrew/apt/Docker/`go install`
publication, SBOMs, attestations or signatures, pre-release tags, automatic version bumps,
changelog files, Hugo in the smoke test, running the workflow, tagging and merging.

## Open decisions (recommendations)

1. Amend constitution 7 to name this workflow beside Pages: recommended, owner's change.
2. **Default: a new version, `v0.1.1`.** sum.golang.org holds a checksum record for
   `github.com/wstein/sheetmill@v0.1.0` and proxy.golang.org serves it as `05eba4d`. If the owner
   re-creates `v0.1.0`, `go install ...@v0.1.0` will resolve the old tree permanently and a direct
   fetch of the new tag (for example `GOPROXY=direct`) will fail with a checksum mismatch.
   Release binaries are unaffected (they stamp the version themselves). The docs present
   `<version>`/`v0.1.1` as the default and the `v0.1.0` retag as a documented, discouraged
   alternative; the choice is the owner's and the workflow serves either.
3. Add build provenance attestations later: recommended after the first release.
4. Decided (owner, 2026-10-17): no required-reviewer rule on the `release` environment. The
   job keeps `environment: release` (GitHub creates it unprotected). Without a reviewer rule any
   owner tag push that passes the checks publishes immediately (Known limitations).
5. Make a `v*` tag ruleset (block deletion/update after release): recommended.
6. Confirm (pending, not granted) that the 2026-10-17 quote is a standing approval for every later `vX.Y.Z` tag
   pushed by the owner (this spec's reading), not for v0.1.0 only. If v0.1.0 only, change
   the `AGENTS.md` line; each later release then needs a new approval.
7. Pull request dry runs: a `pull_request` trigger would exercise the pipeline on every change
   to it but builds binaries on the PR; add it only if the owner approves that.
