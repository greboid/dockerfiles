#!/bin/sh
# Fetch the published installer image and extract its rootfs.
set -eu

dist="${1:-dist}"
mkdir -p "$dist"
rootfs="$dist/rootfs-installer"

if [ -n "${OSIMAGE_INSTALLER_TAR:-}" ]; then
    test -s "$OSIMAGE_INSTALLER_TAR" || { echo "missing $OSIMAGE_INSTALLER_TAR (OSIMAGE_INSTALLER_TAR)" >&2; exit 1; }
    crane export - "$rootfs.tar" < "$OSIMAGE_INSTALLER_TAR"
else
    ref="${OSIMAGE_INSTALLER_REF:-git.mouse-lake.ts.net/osimage/installer:latest}"
    echo "== fetching the installer image ($ref)"
    crane export --platform linux/amd64 "$ref" "$rootfs.tar"
fi

rm -rf "$rootfs"
mkdir -p "$rootfs"
# dev/ nodes need root to create and are not shipped in the squashfs
tar -xf "$rootfs.tar" -C "$rootfs" --numeric-owner --exclude='dev/*'
rm -f "$rootfs.tar"
echo "installer rootfs: $rootfs"
