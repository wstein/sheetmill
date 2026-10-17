#!/bin/sh
# SPDX-License-Identifier: EUPL-1.2
#
# Release notes for a tag on stdout (specs/019-release-workflow, FR-6).
#
# Usage: sh release-notes.sh TAG COMMIT GO_VERSION
#
# Run in a repository with full history and tags. The change list is the non-merge
# commit subjects between the nearest earlier strict release tag (by version order)
# and COMMIT, or "First release." when there is none.

set -eu

[ "$#" -eq 3 ] || {
	printf 'usage: sh release-notes.sh TAG COMMIT GO_VERSION\n' >&2
	exit 2
}
tag=$1
commit=$2
gover=$3
plain=${tag#v}

prev=$(git tag --merged "$commit^" --list 'v*' --sort=-version:refname |
	grep -E '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$' | grep -vx "$tag" | head -n 1 || true)

cat <<TEXT
SheetMill $tag

Commit: $commit
Built with Go $gover, CGO disabled, -trimpath, version stamped as $tag.
Platforms: linux/amd64, linux/arm64, darwin/arm64.
Archives: smill_${plain}_<os>_<arch>.tar.gz, each with smill, LICENSE and README.md.

Verify a download in the directory that holds the archives and SHA256SUMS:

    sha256sum --check --ignore-missing SHA256SUMS
    # macOS, for one archive: grep ' ARCHIVE\$' SHA256SUMS | shasum -a 256 --check

then extract and run \`smill version\`, which prints \`smill $tag\`.

The binaries are not signed or notarized. SHA256SUMS detects a damaged or wrong
download; it does not prove who published the release.

TEXT
if [ -n "$prev" ]; then
	printf '## Changes since %s\n\n' "$prev"
	git log --no-merges --format='- %s (%h)' "$prev..$commit"
else
	printf '## Changes\n\nFirst release.\n'
fi
