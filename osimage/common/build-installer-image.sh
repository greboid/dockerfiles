#!/bin/sh
# Build the unsealed installer image.
set -eu

img_def="${1:?usage: build-installer-image.sh <apko-file> <name> [dist]}"
name="${2:?usage: build-installer-image.sh <apko-file> <name> [dist]}"
out="${3:-dist}"

test -f "$img_def" || { echo "missing image definition $img_def" >&2; exit 1; }
mkdir -p "$out"

echo '== apko: build installer image'
# shellcheck disable=SC2086
SOURCE_DATE_EPOCH=0 apko build ${APKO_EXTRA_REPO_ARGS:-} \
  "$img_def" \
  "localhost/$name:dev" "$out/$name.tar"

echo '== extract installer tree'
rootfs="$out/rootfs-installer"
rm -rf "$rootfs"
mkdir -p "$rootfs"
crane export - "$rootfs.tar" < "$out/$name.tar"
# dev/ nodes need root to create and are not shipped in the squashfs
tar -xf "$rootfs.tar" -C "$rootfs" --numeric-owner --exclude='dev/*'
rm -f "$rootfs.tar"

echo "built $out/$name.tar (extracted: $rootfs)"
