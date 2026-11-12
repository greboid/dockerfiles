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
rm -rf "$out/rootfs" "$out/uki-diagnostics"
rm -f "$out/$name.tar"

echo "== apko: build base image"
# shellcheck disable=SC2086
SOURCE_DATE_EPOCH=0 apko build ${APKO_EXTRA_REPO_ARGS:-} "$img_def" "localhost/$name:dev" "$out/$name.tar"

echo '== seal host UKI'
"$common_dir/seal.sh" "$out/$name.tar" "$out/$name.tar" "$out/rootfs"

echo '== bootc container lint (fatal)'
bootc container lint --rootfs "$out/rootfs"

echo "built $out/$name.tar"
