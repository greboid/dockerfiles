#!/bin/sh
# Seal a host UKI onto an osimage bootc image.
set -eu

base="${1:?usage: seal.sh <base.tar> <out.tar> [rootfs-dir]}"
out="${2:?usage: seal.sh <base.tar> <out.tar> [rootfs-dir]}"
rootfs="${3:-}"

case "$(id -u)" in
    0) ;;
    *) echo "seal.sh: must run as real root (device nodes, capability xattrs)" >&2; exit 1 ;;
esac

out=$(readlink -f "$out")
base=$(readlink -f "$base")
out_dir=$(dirname "$out")
mkdir -p "$out_dir"

# Under out_dir so the final tree can be moved to rootfs-dir without a copy.
work="$out_dir/.seal-work"
rm -rf "$work"
mkdir -p "$work/out"
trap 'rm -rf "$work"' EXIT
tree="$work/bundle/rootfs"

echo "== seal: unpacking $base"
skopeo copy -q "docker-archive:$base" "oci:$work/oci:base"
umoci unpack --image "$work/oci:base" "$work/bundle"

# run/ and tmp/ are tmpfs at boot and bootc lint rejects leftover content.
for d in run tmp; do
    if [ -d "$tree/$d" ]; then find "$tree/$d" -mindepth 1 -delete; fi
done

rm -rf "$tree/boot/EFI/Linux"
mkdir -p "$tree/boot/EFI/Linux"

# umoci repack adds every path whose mtime changed; pin them so the layer is reproducible.
pin_mtimes() {
    for p in "$tree" "$tree/run" "$tree/tmp" "$tree/boot" "$tree/boot/EFI" "$tree/boot/EFI/Linux" "$@"; do
        if [ -e "$p" ]; then touch -h -d @0 "$p"; fi
    done
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
cp "$work/out/$kver.efi" "$tree/boot/EFI/Linux/$kver.efi"
pin_mtimes "$tree/boot/EFI/Linux/$kver.efi"
umoci repack --image "$work/oci:sealed" \
    --history.created 1970-01-01T00:00:00Z \
    --history.created_by 'seal.sh: host UKI' \
    "$work/bundle"
# docker-archive refuses to overwrite; base may be the same file as out
rm -f "$out"
skopeo copy -q "oci:$work/oci:sealed" "docker-archive:$out"

echo "== seal: verifying final archive"
skopeo copy -q "docker-archive:$out" "oci:$work/oci-final:final"
umoci unpack --image "$work/oci-final:final" "$work/final"
final="$work/final/rootfs"
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
