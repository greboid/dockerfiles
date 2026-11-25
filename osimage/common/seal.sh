#!/bin/sh
# Seal a host UKI onto an osimage bootc image.
set -eu

usage='usage: seal.sh <base-oci-layout> <out-oci-layout> [rootfs-dir]'
base="${1:?$usage}"
out="${2:?$usage}"
rootfs="${3:-}"
# archive: unpack the written layout and check its digests; tree: check the sealed tree only.
verify="${SEAL_VERIFY:-archive}"
case "$verify" in
    archive|tree) ;;
    *) echo "seal.sh: SEAL_VERIFY must be archive or tree, got $verify" >&2; exit 1 ;;
esac

case "$(id -u)" in
    0) ;;
    *) echo "seal.sh: must run as real root (device nodes, capability xattrs)" >&2; exit 1 ;;
esac

test -f "$base/index.json" || { echo "seal.sh: $base is not an OCI layout" >&2; exit 1; }
base=$(readlink -f "$base")
mkdir -p "$out"
out=$(readlink -f "$out")
out_dir=$(dirname "$out")
mkdir -p "$out_dir"

# Under out_dir so the final tree can be moved to rootfs-dir without a copy.
work="$out_dir/.seal-work"
rm -rf "$work"
mkdir -p "$work/out" "$work/oci/blobs/sha256" "$work/whiteouts"
trap 'rm -rf "$work"' EXIT
tree="$work/rootfs"
mkdir "$tree"

# Private copy of the layout (blobs hardlinked) so umoci can tag and extend it.
cp "$base/oci-layout" "$work/oci/"
for f in "$base"/blobs/sha256/*; do
    ln "$f" "$work/oci/blobs/sha256/" 2>/dev/null || cp "$f" "$work/oci/blobs/sha256/"
done
jq -e '.manifests | length == 1' "$base/index.json" >/dev/null ||
    { echo "seal.sh: expected exactly one manifest in $base" >&2; exit 1; }
jq '.manifests[0].annotations["org.opencontainers.image.ref.name"] = "base"' \
    "$base/index.json" > "$work/oci/index.json"

manifest="$work/oci/blobs/sha256/$(jq -r '.manifests[0].digest | ltrimstr("sha256:")' "$base/index.json")"
[ "$(jq '.layers | length' "$manifest")" -eq 1 ] ||
    { echo "seal.sh: expected a single-layer base image" >&2; exit 1; }
layer="$work/oci/blobs/sha256/$(jq -r '.layers[0].digest | ltrimstr("sha256:")' "$manifest")"
case "$(jq -r '.layers[0].mediaType' "$manifest")" in
    *.tar+gzip) decompress='gzip -dc' ;;
    *.tar) decompress='cat' ;;
    *) echo "seal.sh: unsupported layer media type" >&2; exit 1 ;;
esac

echo "== seal: extracting $base"
$decompress "$layer" |
    tar -x -p --xattrs --xattrs-include='*' --numeric-owner -C "$tree"

# Remove a directory's contents from the tree and record whiteouts for the UKI layer.
clear_dir() {
    [ -d "$tree/$1" ] || return 0
    find "$tree/$1" -mindepth 1 -maxdepth 1 | while IFS= read -r p; do
        mkdir -p "$work/whiteouts/$1"
        : > "$work/whiteouts/$1/.wh.$(basename "$p")"
        rm -rf "$p"
    done
}
# run/ and tmp/ are tmpfs at boot and bootc lint rejects leftover content.
clear_dir run
clear_dir tmp
clear_dir boot/EFI/Linux
mkdir -p "$tree/boot/EFI/Linux"

# Directories carried in the UKI layer; their mtimes are pinned so the layer is reproducible.
layer_dirs="boot boot/EFI boot/EFI/Linux"
for d in run tmp; do
    if [ -d "$tree/$d" ]; then layer_dirs="$d $layer_dirs"; fi
done
pin_mtimes() {
    touch -h -d @0 "$tree"
    for p in $layer_dirs "$@"; do touch -h -d @0 "$tree/$p"; done
}
pin_mtimes

kver=$(ls "$tree/usr/lib/modules" | head -1)
[ -n "$kver" ] || { echo "seal.sh: no kernel in tree" >&2; exit 1; }
[ "$(ls "$tree/usr/lib/modules" | wc -l)" -eq 1 ] ||
    { echo "seal.sh: multiple kernels in tree" >&2; exit 1; }
echo "== seal: kernel $kver"

set -- \
    --os-release @etc/os-release \
    --stub usr/lib/systemd/boot/efi/linuxx64.efi.stub \
    --output "$work/out/$kver.efi"
if [ -n "${SECUREBOOT_PRIVATE_KEY:-}" ] && [ -n "${SECUREBOOT_CERTIFICATE:-}" ]; then
    echo '== seal: signing (systemd-sbsign)'
    set -- "$@" \
        --signtool systemd-sbsign \
        --secureboot-private-key "$SECUREBOOT_PRIVATE_KEY" \
        --secureboot-certificate "$SECUREBOOT_CERTIFICATE"
else
    echo '== seal: no SECUREBOOT_PRIVATE_KEY/CERTIFICATE; building unsigned'
fi

echo "== seal: bootc container ukify"
bootc container ukify \
    --rootfs "$tree" \
    --write-dumpfile-to "$work/out/$kver.dump" \
    -- "$@"
test -s "$work/out/$kver.efi"

echo "== seal: appending UKI layer"
efi="boot/EFI/Linux/$kver.efi"
cp "$work/out/$kver.efi" "$tree/$efi"
pin_mtimes "$efi"
# shellcheck disable=SC2086
(cd "$tree" && tar -c --format=gnu --numeric-owner --no-recursion -f "$work/uki-layer.tar" $layer_dirs "$efi")
whiteouts=$(cd "$work/whiteouts" && find . -type f | sed 's|^\./||' | sort)
if [ -n "$whiteouts" ]; then
    (cd "$work/whiteouts" && echo "$whiteouts" | while IFS= read -r w; do touch -d @0 "$w"; done &&
        echo "$whiteouts" | tar -r --format=gnu --numeric-owner --owner=0 --group=0 --no-recursion \
            -f "$work/uki-layer.tar" -T -)
fi
umoci raw add-layer --image "$work/oci:base" --tag sealed \
    --history.created 1970-01-01T00:00:00Z \
    --history.created_by 'seal.sh: host UKI' \
    "$work/uki-layer.tar"

# umoci omits the manifest mediaType; set it and make the sealed image the layout's only entry.
layout="$work/oci"
sealed=$(jq -c '.manifests[] | select(.annotations["org.opencontainers.image.ref.name"] == "sealed")' "$layout/index.json")
old_manifest="$layout/blobs/sha256/$(echo "$sealed" | jq -r '.digest | ltrimstr("sha256:")')"
jq -c '{schemaVersion, mediaType: "application/vnd.oci.image.manifest.v1+json"} + .' "$old_manifest" > "$work/manifest.json"
mdigest=$(sha256sum "$work/manifest.json" | cut -d' ' -f1)
msize=$(wc -c < "$work/manifest.json")
mv "$work/manifest.json" "$layout/blobs/sha256/$mdigest"
jq -c --arg d "sha256:$mdigest" --argjson s "$msize" --argjson p "$(jq -c '.manifests[0].platform // null' "$base/index.json")" '
    {schemaVersion: 2, mediaType: "application/vnd.oci.image.index.v1+json",
     manifests: [{mediaType: "application/vnd.oci.image.manifest.v1+json", digest: $d, size: $s}
                 + (if $p then {platform: $p} else {} end)
                 + {annotations: {"org.opencontainers.image.ref.name": "latest"}}]}' \
    "$layout/index.json" > "$work/index.json"
mv "$work/index.json" "$layout/index.json"
umoci gc --layout "$layout"
rm -rf "$out"
mv "$layout" "$out"

if [ "$verify" = archive ]; then
    echo "== seal: verifying final layout"
    umoci unpack --image "$out:latest" "$work/final"
    final="$work/final/rootfs"
else
    echo "== seal: verifying sealed tree (archive round-trip skipped)"
    final="$tree"
fi
d1=$(bootc container compute-composefs-digest "$final")
d2=$(bootc container compute-composefs-digest --erofs-version=v2 "$final")
for d in "$d1" "$d2"; do
    case "$d" in
        [0-9a-f]*) [ ${#d} -eq 128 ] || { echo "seal.sh: bad digest: $d" >&2; exit 1; } ;;
        *) echo "seal.sh: bad digest: $d" >&2; exit 1 ;;
    esac
done

objcopy -O binary --only-section=.cmdline \
    "$work/out/$kver.efi" "$work/out/cmdline.bin"
cmdline=$(tr -d '\0' < "$work/out/cmdline.bin")
case "$cmdline" in
    *"composefs.digest=v1-sha512-12:$d1 "*) ;;
    *)
        echo "seal.sh: embedded cmdline does not carry the computed V1 digest" >&2
        echo "  cmdline: $cmdline" >&2
        echo "  computed v1: $d1" >&2
        exit 1
        ;;
esac
case "$cmdline" in
    *" composefs=$d2"*) ;;
    *)
        echo "seal.sh: embedded cmdline does not carry the computed V2 digest" >&2
        echo "  cmdline: $cmdline" >&2
        echo "  computed v2: $d2" >&2
        exit 1
        ;;
esac

nukis=$(find "$final/boot/EFI/Linux" -name '*.efi' | wc -l)
[ "$nukis" -eq 1 ] || { echo "seal.sh: expected exactly one UKI, found $nukis" >&2; exit 1; }

diag="$out_dir/uki-diagnostics"
mkdir -p "$diag"
cp "$work/out/$kver.efi" "$diag/$kver.efi"
cp "$work/out/$kver.dump" "$diag/$kver.dump" 2>/dev/null || true
cp "$work/out/cmdline.bin" "$diag/cmdline.bin"
{
    echo "kver: $kver"
    echo "v1: $d1"
    echo "v2: $d2"
    echo "cmdline: $cmdline"
} > "$diag/summary.txt"

if [ -n "$rootfs" ]; then
    rm -rf "$rootfs"
    mv "$final" "$rootfs"
fi

echo "seal: UKI $kver.efi sealed (v1 $d1)"
