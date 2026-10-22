# Releasing SheetMill

Releases are made by CI from a version tag. This page is the owner procedure; the design is
in [specs/019-release-workflow](../specs/019-release-workflow/spec.md). Agents and
contributors never create, move or delete tags and never start the workflow.

## Approval

There is no standing approval. Every tag and release, including later ones, needs Werner
Stein's explicit approval each time; an approval for one tag (for example `v0.1.1`) covers only
that tag. The approval is the owner's own push of the tag; it is not written per version
into `AGENTS.md`. Agents and contributors never create or push tags on that basis.

## What a tag does

Pushing an annotated tag `vMAJOR.MINOR.PATCH` (for example `v0.1.1`; pre-release suffixes
such as `-rc1` start nothing) runs `.github/workflows/release.yml`:

1. **verify**: the tag is annotated, names the commit being built, is an ancestor of `main`
   and was pushed by the repository owner (the same check guards manual dry runs); the `check` workflow succeeded for that commit; no
   release exists for the tag; `.go-version` equals `go.mod`; `make check`, `make mod-verify`
   and `make test-assets` pass.
   The CI guard retries its `head_sha` query up to five times and then filters the newest 500
   successful push runs of `check.yml` by commit client side, because the Actions API has answered
   that query alternately with two results and none; every run is re-checked for the commit, a
   success conclusion and the push event, and a commit older than those 500 runs must re-run
   `check` first.
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

## One-time setup (owner, optional)

- The `release` environment needs no setup: the owner decided (2026-10-17) not to configure a
  required-reviewer rule. GitHub creates the environment unprotected on first use, so the
  pushed tag alone releases; the workflow's own checks are the control.
- Rulesets: protect `v*` tags against deletion and updates once the release exists.
- Repository variable `SHEETMILL_RELEASE_DRAFT=true` makes the next release a draft that you
  publish by hand (default: published immediately).

## Release a version (owner only)

The default path is a **new** version, `v0.1.1` in the first release (any unused
`vMAJOR.MINOR.PATCH` works; the examples below use `<version>`). The existing annotated tag
`v0.1.0` (`05eba4d`, 2026-10-16, no release) is left alone.

1. Review and merge the workflow pull request. The workflow file must be on `main`, because a
   tag push runs the workflow from the tagged commit.
2. Choose the `main` commit `<sha>`: it contains features 017 and 018 and `release.yml`, and
   its `check` run is green (re-run `check` on it if it was cancelled).
3. Optional, recommended once: run a dry run (see the end of this page) on `main`.
4. Create and push the annotated tag:

   ```sh
   git fetch origin main
   git tag -a <version> -m "Release smill <version>" <sha>
   git push origin refs/tags/<version>
   ```

5. Watch the `release` run under Actions. Publish starts as soon as build and smoke pass. If
   a reviewer rule is ever added to the environment, approve within 24 hours: the build
   artifacts expire after one day (`retention-days: 1`), and a later approval fails at the
   download (see "When something goes wrong").

## Discouraged alternative: re-create v0.1.0

The owner's 2026-10-17 wording named `v0.1.0`. Re-creating that tag **will** have these
effects, so it is not the default: sum.golang.org already holds a checksum record for
`github.com/wstein/sheetmill@v0.1.0` and proxy.golang.org serves `v0.1.0` as `05eba4d`
(`https://proxy.golang.org/github.com/wstein/sheetmill/@v/v0.1.0.info`). After a retag,
`go install github.com/wstein/sheetmill/cmd/smill@v0.1.0` resolves the old tree
permanently, and a direct fetch of the new tag (for example `GOPROXY=direct`) fails with a
`SECURITY ERROR` checksum mismatch. The release binaries are not affected. If you accept this:

```sh
git tag -d v0.1.0
git push origin :refs/tags/v0.1.0      # deleting a tag does not start the workflow
git tag -a v0.1.0 -m "Release smill v0.1.0" <sha>
git push origin refs/tags/v0.1.0
```

and then continue with step 5 above.

## Verify the release

(`<number>` is the version without the leading `v`.)

```sh
mkdir /tmp/smill-check && cd /tmp/smill-check
gh release download <version> --repo wstein/sheetmill
sha256sum --check --ignore-missing SHA256SUMS
tar -xzf smill_<number>_linux_amd64.tar.gz && ./smill_<number>_linux_amd64/smill version   # smill <version>
```

On the Apple silicon Mac mini, download `smill_<number>_darwin_arm64.tar.gz` and `SHA256SUMS`, run
`grep ' smill_<number>_darwin_arm64.tar.gz$' SHA256SUMS | shasum -a 256 --check`, extract, and run
`smill version` and `smill check examples/sheets` from a checkout. Downloads made with `curl`
or `gh` are not quarantined; a browser download is, and Gatekeeper may refuse the unsigned
binary (clear it with `xattr -d com.apple.quarantine <file>` only if you trust the checksum).

## Known limitations

Without a required-reviewer rule on the `release` environment, any tag push by the owner that
passes the checks publishes immediately; there is no pause between the push and the public
release. The remaining controls are the owner-only actor checks, the annotated-tag, ancestor,
version and green-`check` verification, the refusal to publish when a release or draft exists,
and the rule never to delete and re-create a published version's tag. Nothing is signed.

## Maintaining the workflow

`internal/releaseworkflow/testdata/release.golden.json` pins the whole parsed workflow. It is
a change detector, not a safety proof: running the tests with `UPDATE_RELEASE_GOLDEN=1`
regenerates it (the run then fails on purpose and is refused when `CI` is set), so read the
diff. The real checks are the semantic rules and the mutation table in `workflow_test.go`.
An action bump needs the golden, the action pin and allow-list in the test, and a few mutation
rows; mutation rows match raw workflow text, so harmless reformatting of `release.yml` makes
them fail with "HARNESS ERROR, NOT A KILL" until the rows are updated.

## When something goes wrong

- **A job fails before publish**: nothing is public. If the cause is transient, re-run the
  failed jobs with the same tag and commit; if the build artifacts have expired (one day),
  use Re-run all jobs. Check `gh release list` first: a release or draft for the tag blocks
  publish. If the tagged commit itself must change, do **not** delete and re-create the tag:
  the module proxy fetches any tag on demand (for example through `@latest`) and keeps the
  first content permanently, which is the same hazard as retagging `v0.1.0`. Tag the next
  patch version on the fixed commit instead.
- **Publish fails midway**: run `gh release view <version>`. If a release or draft exists with
  missing assets, delete it with `gh release delete <version> --yes` (the tag stays) and re-run
  the workflow. Do not edit assets by hand.
- **A bad release is public (yank)**: `gh release delete <version> --yes`, add the reason to the
  next release notes, and publish a fixed higher version rather than moving the tag again.
  This assumes GitHub immutable releases are off; with them on, a published release and its
  tag cannot be changed or reused, so always fix forward.
- **Dry run**: Actions, release, Run workflow, on a branch (owner only). It builds and
  smoke-tests with the label `v0.0.0-dryrun` (any `vX.Y.Z-label` is accepted) and never
  publishes; artifacts are deleted after one day. Run one before the first real tag. A dry run never
  publishes, but in a public repository its workflow artifacts (unsigned binaries labelled
  `-dryrun`) can be downloaded by any signed-in GitHub user until they expire.
