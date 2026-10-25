# Research: 019 release workflow

Ratings: 5 best. Chosen option in bold. Facts marked "checked" were verified on
2026-10-17 from this worktree or with `git ls-remote`/`curl`; others are from GitHub
documentation as the author knows it and are listed under "Unverified".

## D1 Trigger and pre-releases

| Option | Safety | Simplicity | Fit | Note |
| --- | --- | --- | --- | --- |
| **Tag push `v[0-9]+.[0-9]+.[0-9]+` (whole ref must match)** | 5 | 5 | 5 | `v1.2.3-rc1` triggers nothing; no pre-release logic |
| `v*` plus script check | 3 | 4 | 3 | Every odd tag starts a run that must fail |
| Release-created event | 2 | 2 | 2 | Creates the release before checks |
| Pre-release suffix support | 3 | 2 | 2 | Needs prerelease flag, ordering rules; no demand |

The strict regex is checked again in `verify` because a glob is easier to widen by
accident than a regex; a test pins both.

## D2 Verify: how to prove "green"

| Option | Result |
| --- | --- |
| Query `check` push run for the commit via `gh api` | **Chosen**, needs `actions: read`. Reuses Hugo, race and fuzz evidence at no cost. Fails closed with an instruction. `check.yml` has `cancel-in-progress: true` per ref, so a superseded `main` run is canceled and not green: tagging an old commit needs `check` re-run on it first. Acceptable, and a feature (the owner tags a verified commit) |
| Re-run everything in release | Duplicates Hugo download and fuzz; slower; rejected |
| Branch protection only | Not enforced for tag pushes; rejected |
| **Also** re-run `make check`, `make mod-verify`, `make test-assets` | Chosen: cheap, and the only gates in dry runs where no `check` evidence exists for the merge ref |

Other checks: annotated tag (a lightweight tag has no tagger, message or date, and the
owner's convention is annotated), ancestor of `origin/main` (rules out a tag on an
unreviewed branch), tag pusher equals the repository owner (the owner's tag is the
approval), no release exists (re-runs after a half-failure need a clean slate; fail
closed on any API error other than 404).

## D3 Runners

| Target | Runner | Why |
| --- | --- | --- |
| build (all) | `ubuntu-24.04` | Go cross-compiles with CGO off; one OS gives GNU `tar` for deterministic archives |
| smoke linux/amd64 | `ubuntu-24.04` | native |
| smoke linux/arm64 | `ubuntu-24.04-arm` | arm64 Linux runners are free for public repositories (unverified for this repo; if the label is unavailable the job queues and then fails, which blocks publish, fail closed) |
| smoke darwin/arm64 | `macos-15` | Apple silicon; `macos-13` (Intel) and darwin/amd64 are end of life |

Versions are pinned (not `-latest`) so an image roll cannot change a release run silently;
`check.yml` uses `ubuntu-latest` for cheap compile checks and is unchanged.
Alternative: build natively on each runner (rejected: three toolchain installs, BSD tar
on macOS, no extra assurance for CGO-free Go).

## D4 CI pipeline update

| Option | Result |
| --- | --- |
| Reusable `workflow_call` of `release.yml` from a `check.yml` job | Rejected: a called workflow may not request more permission than its caller grants; `publish` asks for `contents: write`, which the validator rejects even for a skipped job (unverified, known to fail at startup) |
| New job in `check.yml` duplicating the build/smoke steps | Rejected: two copies drift |
| `pull_request` + `paths` trigger on `release.yml`, dry run | Rejected in review round 1: the draft PR itself would build and upload three binaries, a distribution build needing separate approval (AGENTS.md) |
| **`workflow_dispatch` dry run only, started by the owner** | Chosen: one copy, cannot publish, no unapproved build; open decision 7 keeps the PR trigger as an option |

`check.yml` keeps names and behavior; its `build` matrix is the cheap compile guard on
every push. A parity test ties the two matrices together so they cannot diverge.

## D5 Reproducibility

Go with `CGO_ENABLED=0 -trimpath` and a fixed ldflags value is reproducible for a given
toolchain and module set. VCS stamping stays on (`vcs.revision` is deterministic per
commit; checkout leaves the tree clean). The workflow builds each target twice with an
empty `GOCACHE` for the second and `cmp`s them. This proves no hidden cache or temp-path
dependence on one machine; it does not prove independent reproduction (accepted limitation).
Archives are deterministic from the binary: sorted names, owner 0, `mtime` = commit time,
`gzip -n`; a script test packages twice and compares bytes.

## D6 Smoke test content

`smill version` equals the tag (proves the stamp reached the shipped bytes), `--help`,
`check examples/sheets` (the independently authored sheets, 15 files, checked on this
tree: exit 0), and `build hugo examples/sheets --out X` twice with `diff -r` (checked:
exports 15 sheets without Hugo; "local generation is a separate command"). This exercises
parser, validator and renderer export natively on all three targets without a Hugo
download. Rejected: running Hugo (a 90 MB pinned download per runner for little new signal
about the Go binary; `check.yml` already covers the Hugo path); cross-platform digest
comparison (needs another artifact; deferred).

## D7 Attestations and signing

| Option | Result |
| --- | --- |
| **SHA256SUMS only** | Chosen. Detects corruption and wrong-file mixups. Does not prove origin (served from the same release) |
| `actions/attest-build-provenance` | Deferred (open decision 3). Needs `id-token: write` and `attestations: write` on the one job that holds `contents: write`, a new action pin and a consumer-side `gh attestation verify` habit. Valuable; not needed to ship a first release |
| cosign / minisign / GPG | Rejected: key management, secrets |
| macOS codesign/notarization | Rejected: needs an Apple developer account and secrets |

The release notes and docs say plainly that nothing is signed.

## D8 Draft or published

| Option | Result |
| --- | --- |
| **Publish directly; the owner's tag is the approval, a protected `release` environment is an optional extra pause** | Chosen: the owner asked for a CI release. A missing environment is created unprotected by GitHub, so the workflow does not rely on protection (publish re-checks the actor) |
| Draft first | Adds a second manual step for the same decision; draft releases do not appear to users but a tag already exists; useful only when the owner wants to edit notes |
| Both via switch | **Also provided**: repository variable `SHEETMILL_RELEASE_DRAFT=true` adds `--draft`; no workflow edit, no secret |

Immutable releases (a GitHub repository setting, unverified) would make a published
release permanent; `docs/releasing.md` states the yank procedure assumes they are off.

## D9 Retagging `v0.1.0` and the Go module proxy

Checked: `git ls-remote --tags origin` shows annotated `v0.1.0` (`603279a`) peeling to
`05eba4d`, which is an ancestor of `origin/main` (`86fc231`) with 59 later commits including
017 and 018 changes. Checked: `https://proxy.golang.org/github.com/wstein/sheetmill/@v/list`
returns `v0.1.0` and `.info` reports Hash `05eba4d`, time 2026-10-16T15:58:27Z. The proxy
and the checksum database (not queried: a lookup would record the version) keep the first content for a version permanently; deleting and
re-creating the tag does not change what `go install github.com/wstein/sheetmill/cmd/smill@v0.1.0`
resolves. Consequences: release binaries (this feature) carry the new tree; `@v0.1.0`
module consumers get the old tree; if a consumer's `go.sum` records the proxy hash, a
local checkout of the new tag mismatches it. The workflow stamps the version into binaries
so the binary is correct either way. This is the owner's decision (open decision 2) and the
risk is stated in `docs/releasing.md`. Tag deletion does not fire `push` workflows, so the
procedure is safe in that respect (unverified).

## D10 Actions and pins

Reused from this repository: `actions/checkout@fbc6f3992d24b796d5a048ff273f7fcc4a7b6c09` (v5),
`actions/setup-go@924ae3a1cded613372ab5595356fb5720e22ba16` (v6).
New, SHA resolved with `git ls-remote https://github.com/actions/<name> 'refs/tags/v*'`
on 2026-10-17 (lightweight tags, so the listed SHA is the commit):

- `actions/upload-artifact@cf430e030ddbb5b0abf93d22962f4752f3646cd9` (v7.0.2)
- `actions/download-artifact@9000827ccba6bdab643e8b6fd33ac0654aef8333` (v8.0.2)

Not verified: that these two are free of known advisories, and artifact v7/v8 pairing
(both use the v4 artifact backend). `gh` is used from the runner image (preinstalled on
GitHub-hosted runners, unverified); no release action from the marketplace is used.
`gh` and `curl` are not downloaded by the workflow.

## D11 Archive layout and naming

`smill_<version>_<os>_<arch>.tar.gz` with `<version>` without `v` (common convention, avoids
`smill_v0.1.0_`); a top directory so extraction never scatters files; `tar.gz` on all
three targets (no Windows zip). Alternative bare binaries: rejected, no LICENSE (EUPL
requires the license to travel with the work) and no checksum-friendly layout.

## D12 Release notes

Fixed template plus `git log --no-merges --format='- %s (%h)'` for `<prev>..<tag>`, where
`<prev>` is the nearest earlier `v*` tag reachable from the tag; none means "First release."
(the first release would otherwise list the whole history). Rejected: `--generate-notes`
(pulls PR titles, wording not reviewable in the repo), a CHANGELOG file (new maintenance).

## Unverified

Runner label availability, GitHub's handling of an annotated tag by `actions/checkout`
(the workflow fetches the tag explicitly to avoid depending on it), `gh release create`
upload atomicity, environment/`vars` behavior, the 404 body format of the releases API
(fail-closed, so a different body blocks a release rather than allowing one), and the
whole tag path before the first real tag.
