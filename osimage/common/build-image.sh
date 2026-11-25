#!/bin/sh
# Build a bootc OS image and seal a host UKI onto it.
set -eu

img_def="${1:?usage: build-image.sh <apko-file> <name> [dist]}"
name="${2:?usage: build-image.sh <apko-file> <name> [dist]}"
out="${3:-dist}"
common_dir="$(cd "$(dirname "$0")" && pwd)"

test -f "$img_def" || { echo "missing image definition $img_def" >&2; exit 1; }
if ! command -v bootc >/dev/null 2>&1; then
    echo 'bootc missing: install the org bootc package (seal and container lint)' >&2
    exit 1
fi
mkdir -p "$out"
base="$out/.base-oci"
rm -rf "$out/rootfs" "$out/uki-diagnostics" "$base"
rm -rf "$out/$name.oci" "$out/$name.tar"
trap 'rm -rf "$base"' EXIT

echo "== apko: build base image"
# apko writes an OCI layout when the output is an existing directory.
mkdir -p "$base"
# shellcheck disable=SC2086
SOURCE_DATE_EPOCH=0 apko build ${APKO_EXTRA_REPO_ARGS:-} ${APKO_CACHE_DIR:+--cache-dir "$APKO_CACHE_DIR"} "$img_def" "localhost/$name:dev" "$base/"

echo '== seal host UKI'
"$common_dir/seal.sh" "$base" "$out/$name.oci" "$out/rootfs"

echo '== bootc container lint (fatal)'
# /etc is not persisted (composefs [etc] mount = "none"), so users cannot drift
bootc container lint --rootfs "$out/rootfs" --skip sysusers

echo "built $out/$name.oci"
