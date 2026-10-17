#!/bin/sh
# SPDX-License-Identifier: EUPL-1.2
#
# Package one release binary (specs/019-release-workflow, FR-5).
#
# Usage: SOURCE_DATE_EPOCH=N sh release-package.sh VERSION OS ARCH BINARY OUTDIR
#
# Run in the repository root (LICENSE and README.md are read from there) with GNU tar.
# Writes OUTDIR/smill_<version without v>_<os>_<arch>.tar.gz and its .sha256 sidecar.
# The archive holds one top directory with smill (0755), LICENSE and README.md (0644),
# in sorted order, owner 0, mtime SOURCE_DATE_EPOCH, no gzip name or time, so the same
# inputs always give the same bytes.

set -eu

fail() {
	printf 'release-package: %s\n' "$*" >&2
	exit 1
}

[ "$#" -eq 5 ] || {
	printf 'usage: SOURCE_DATE_EPOCH=N sh release-package.sh VERSION OS ARCH BINARY OUTDIR\n' >&2
	exit 2
}
version=$1
os=$2
arch=$3
binary=$4
out=$5

printf '%s\n' "$version" | grep -Eq '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$' ||
	fail "version must look like v1.2.3: $version"
printf '%s_%s\n' "$os" "$arch" | grep -Eq '^[a-z0-9]+_[a-z0-9]+$' || fail "bad os or arch: $os $arch"
: "${SOURCE_DATE_EPOCH:?release-package: SOURCE_DATE_EPOCH is required}"
printf '%s\n' "$SOURCE_DATE_EPOCH" | grep -Eq '^[0-9]+$' || fail "SOURCE_DATE_EPOCH must be digits"
[ -f "$binary" ] || fail "binary not found: $binary"
[ -f LICENSE ] || fail "LICENSE not found in $(pwd)"
[ -f README.md ] || fail "README.md not found in $(pwd)"
tar --version 2>/dev/null | grep -q 'GNU tar' || fail "GNU tar is required"

name=smill_${version#v}_${os}_${arch}
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

mkdir "$work/$name"
chmod 0755 "$work/$name"
cp "$binary" "$work/$name/smill"
cp LICENSE README.md "$work/$name/"
chmod 0755 "$work/$name/smill"
chmod 0644 "$work/$name/LICENSE" "$work/$name/README.md"

mkdir -p "$out"
archive=$out/$name.tar.gz
tar --sort=name --owner=0 --group=0 --numeric-owner --mtime="@$SOURCE_DATE_EPOCH" \
	--format=ustar -C "$work" -cf - "$name" | gzip -n -9 >"$archive"
(cd "$out" && sha256sum "$name.tar.gz" >"$name.tar.gz.sha256")
printf '%s\n' "$archive"
