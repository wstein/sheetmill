# Releasing SheetMill

Releases are made by CI from a version tag. This page is the owner procedure; the design is
in [specs/019-release-workflow](../specs/019-release-workflow/spec.md). Agents and
contributors never create, move or delete tags and never start the workflow.

## What a tag does

Pushing an annotated tag `vMAJOR.MINOR.PATCH` (for example `v0.1.0`; pre-release suffixes
such as `-rc1` start nothing) runs `.github/workflows/release.yml`:

1. **verify**: the tag is annotated, names the commit being built, is an ancestor of `main`
   and was pushed by the repository owner; the `check` workflow succeeded for that commit; no
   release exists for the tag; `.go-version` equals `go.mod`; `make check`, `make mod-verify`
   and `make test-assets` pass.
2. **build**: `linux/amd64`, `linux/arm64` and `darwin/arm64` with `CGO_ENABLED=0`,
   `-trimpath` and the tag stamped into `smill version`. Each target is built twice (the second
   time with an empty build cache) and the two binaries must be byte-identical. Archives
   `smill_<version>_<os>_<arch>.tar.gz` hold `smill`, `LICENSE` and `README.md`.
3. **smoke**: each archive is checked against its checksum and run on a native runner
   (`ubuntu-24.04`, `ubuntu-24.04-arm`, `macos-15`): `smill version`, `--help`,
   `check examples/sheets` and two identical `build hugo` exports. Hugo is not run.
4. **publish** (only `contents: write` job; environment `release`): re-checks the actor and
   that no release or draft exists, confirms the remote tag still names the verified commit,
   writes `SHA256SUMS`, and runs `gh release create <tag> --verify-tag` with the archives.

Nothing is signed or notarized and there are no attestations. `SHA256SUMS` detects a damaged
download, not a malicious release. Reproducibility is checked on one runner only.

## One-time setup (owner, optional but recommended)

- Settings, Environments, `release`: add yourself as required reviewer so publish waits for
  your approval; allow deployment from tags matching `v*`. Without this the environment is
  created unprotected and the pushed tag alone releases.
- Rulesets: protect `v*` tags against deletion and updates **after** the retag below.
- Repository variable `SHEETMILL_RELEASE_DRAFT=true` makes the next release a draft that you
  publish by hand (default: published immediately).

## Decide first: re-create v0.1.0 or tag v0.1.1

The existing annotated tag `v0.1.0` (`05eba4d`, 2026-10-16, no release) points before the 59 later `main` commits, which carry
features 017 and 018 as shipped. The Go module proxy has already cached `v0.1.0` for
`05eba4d` (`https://proxy.golang.org/github.com/wstein/sheetmill/@v/v0.1.0.info`) and the
checksum database most likely holds it. Re-creating the tag does not change that:
`go install github.com/wstein/sheetmill/cmd/smill@v0.1.0` keeps resolving the old tree, and a
direct fetch of the new tag can fail with a checksum mismatch. The release binaries are not
affected. If you do not want that, use `v0.1.1` and skip the deletion steps below.

## Retag v0.1.0 (owner only)

1. Review and merge the workflow pull request. The workflow file must be on `main`, because a
   tag push runs the workflow from the tagged commit.
2. Choose the `main` commit `<sha>`: it contains features 017 and 018 and `release.yml`, and
   its `check` run is green (re-run `check` on it if it was cancelled).
3. Delete the old tag everywhere (deleting does not start the workflow):

   ```sh
   git tag -d v0.1.0
   git push origin :refs/tags/v0.1.0
   ```

4. Create and push the annotated tag:

   ```sh
   git fetch origin main
   git tag -a v0.1.0 -m "Release smill v0.1.0" <sha>
   git push origin refs/tags/v0.1.0
   ```

5. Watch the `release` run under Actions. Approve the `release` environment if protected.

## Verify the release

```sh
mkdir /tmp/smill-check && cd /tmp/smill-check
gh release download v0.1.0 --repo wstein/sheetmill
sha256sum --check --ignore-missing SHA256SUMS
tar -xzf smill_0.1.0_linux_amd64.tar.gz && ./smill_0.1.0_linux_amd64/smill version   # smill v0.1.0
```

On the Apple silicon Mac mini, download `smill_0.1.0_darwin_arm64.tar.gz` and `SHA256SUMS`, run
`grep ' smill_0.1.0_darwin_arm64.tar.gz$' SHA256SUMS | shasum -a 256 --check`, extract, and run
`smill version` and `smill check examples/sheets` from a checkout. Downloads made with `curl`
or `gh` are not quarantined; a browser download is, and Gatekeeper may refuse the unsigned
binary (clear it with `xattr -d com.apple.quarantine <file>` only if you trust the checksum).

## When something goes wrong

- **A job fails before publish**: nothing is public. Fix the cause, then either re-run the failed
  jobs (same tag, same commit) or, if the tagged commit must change, delete and re-create the
  tag as above. Check `gh release list` first: a release or draft for the tag blocks publish.
- **Publish fails midway**: run `gh release view v0.1.0`. If a release or draft exists with
  missing assets, delete it with `gh release delete v0.1.0 --yes` (the tag stays) and re-run
  the workflow. Do not edit assets by hand.
- **A bad release is public (yank)**: `gh release delete v0.1.0 --yes`, add the reason to the
  next release notes, and publish a fixed higher version rather than moving the tag again.
  This assumes GitHub immutable releases are off; with them on, a published release and its
  tag cannot be changed or reused, so always fix forward.
- **Dry run**: Actions, release, Run workflow, on a branch. It builds and smoke-tests with the
  label `v0.0.0-dryrun` and never publishes; artifacts are deleted after one day.
