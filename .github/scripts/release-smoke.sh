#!/bin/sh
# SPDX-License-Identifier: EUPL-1.2
#
# Exercise a shipped release archive on its native host (specs/019-release-workflow, FR-4).
#
# Usage: sh release-smoke.sh VERSION ARCHIVE REPO_ROOT
#
# ARCHIVE is smill_<version>_<os>_<arch>.tar.gz next to its .sha256 sidecar. REPO_ROOT
# supplies examples/sheets. Checks, in order: the checksum; the archive members; that
# `smill version` prints "smill VERSION"; `--help`; `check examples/sheets`; and that
# `build hugo examples/sheets` run twice writes identical, non-empty trees. Hugo is not
# run. Works with POSIX tools on Linux and macOS.

set -eu

fail() {
	printf 'release-smoke: %s\n' "$*" >&2
	exit 1
}

[ "$#" -eq 3 ] || {
	printf 'usage: sh release-smoke.sh VERSION ARCHIVE REPO_ROOT\n' >&2
	exit 2
}
version=$1
archive=$2
root=$3

[ -f "$archive" ] || fail "archive not found: $archive"
[ -d "$root/examples/sheets" ] || fail "no examples/sheets in $root"
archive=$(cd "$(dirname "$archive")" && pwd)/$(basename "$archive")
root=$(cd "$root" && pwd -P)
file=$(basename "$archive")
name=${file%.tar.gz}
[ "$name" != "$file" ] || fail "not a .tar.gz: $file"

# Checksum written at build time (sha256sum on Linux, shasum on macOS).
[ -f "$archive.sha256" ] || fail "checksum sidecar missing: $file.sha256"
if command -v sha256sum >/dev/null 2>&1; then
	(cd "$(dirname "$archive")" && sha256sum -c "$file.sha256" >/dev/null) || fail "checksum mismatch for $file"
else
	(cd "$(dirname "$archive")" && shasum -a 256 -c "$file.sha256" >/dev/null) || fail "checksum mismatch for $file"
fi

# Physical path: smill refuses paths through a symlink (macOS TMPDIR is under /var -> /private/var).
d=$(mktemp -d) || fail "cannot create a temp dir"
trap 'rm -rf "$d"' EXIT
tmp=$(cd "$d" && pwd -P) || fail "cannot resolve the temp dir"
tar -xzf "$archive" -C "$tmp" || fail "cannot extract $file"

members=$(cd "$tmp" && find . -mindepth 1 | LC_ALL=C sort | tr '\n' ' ')
expected="./$name ./$name/LICENSE ./$name/README.md ./$name/smill "
[ "$members" = "$expected" ] || fail "unexpected archive members: $members"
bin=$tmp/$name/smill
[ -x "$bin" ] || fail "smill is not executable"

got=$("$bin" version) || fail "smill version failed"
[ "$got" = "smill $version" ] || fail "smill version printed '$got', want 'smill $version'"
"$bin" --help >/dev/null || fail "smill --help failed"
(cd "$root" && "$bin" check examples/sheets) || fail "smill check examples/sheets failed"
(cd "$root" && "$bin" build hugo examples/sheets --out "$tmp/export-a" >/dev/null) || fail "first export failed"
(cd "$root" && "$bin" build hugo examples/sheets --out "$tmp/export-b" >/dev/null) || fail "second export failed"
files=$(find "$tmp/export-a" -type f | wc -l | tr -d ' ')
[ "$files" -gt 0 ] || fail "export wrote no files"
diff -r "$tmp/export-a" "$tmp/export-b" >/dev/null || fail "export is not deterministic: two runs differ"
printf 'release-smoke: %s ok (%s export files)\n' "$file" "$files"
