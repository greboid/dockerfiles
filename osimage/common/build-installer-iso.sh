#!/bin/sh
# Build the installer ISO: a GPT image (ESP with the live UKI + LiveOS squashfs
# partition) with an El Torito catalog, bootable from optical media or dd'd to USB.
set -eu

export SOURCE_DATE_EPOCH=0

dist="${1:-dist}"
common_dir="$(cd "$(dirname "$0")" && pwd)"
rootfs="$dist/rootfs-installer"
test -d "$rootfs/usr/lib/modules" || { echo "missing $rootfs (run common/fetch-installer-rootfs.sh $dist, or build-installer-image.sh for a local build)" >&2; exit 1; }
test -s "$rootfs/usr/lib/systemd/boot/efi/linuxx64.efi.stub" || {
    echo "missing systemd-efistub in $rootfs" >&2; exit 1; }
test -s "$common_dir/keys/packages.rsa.pub" || {
    echo "missing $common_dir/keys/packages.rsa.pub (package signing key)" >&2; exit 1; }

for tool in ukify systemd-repart apk curl; do
    command -v "$tool" >/dev/null || { echo "$tool not found" >&2; exit 1; }
done

ORG_REPO="${ORG_REPO:-https://git.mouse-lake.ts.net/api/packages/packages/alpine/v3.24/main}"
kver=$(ls "$rootfs/usr/lib/modules")
case "$kver" in
    '' | *[[:space:]]*) echo "expected exactly one kernel in $rootfs/usr/lib/modules" >&2; exit 1 ;;
esac
kernel="$rootfs/usr/lib/modules/$kver/vmlinuz"
test -s "$kernel"

kernel_version=$(awk -F: '
    $1 == "P" { p = $2 }
    p == "osimage-kernel" && $1 == "V" { print $2; exit }
' "$rootfs/lib/apk/db/installed")
test -n "$kernel_version" || { echo "cannot read the osimage-kernel version from $rootfs/lib/apk/db/installed" >&2; exit 1; }

builddir="$dist/iso-installer"
iso="$dist/installer.iso"
rm -rf "$builddir" "$iso"
mkdir -p "$builddir/initramfs-live" "$builddir/tmp"

echo '== installer initramfs (osimage-kernel-initramfs-live package, signature checked)'
curl -fsSL --retry 3 "$ORG_REPO/x86_64/osimage-kernel-initramfs-live-$kernel_version.apk" \
  -o "$builddir/initramfs-live.apk"
apk extract --keys-dir "$common_dir/keys" --no-chown \
  --destination "$builddir/initramfs-live" "$builddir/initramfs-live.apk"
live_initramfs="$builddir/initramfs-live/usr/lib/modules/$kver/initramfs-install.img"
test -s "$live_initramfs"

echo '== live UKI (ukify; kernel + live initramfs + installer cmdline)'
cmdline="systemd.volatile=overlay systemd.mask=boot.automount systemd.mask=boot.mount console=tty0 console=ttyS0,115200 init=/usr/lib/systemd/systemd"
set -- \
  --efi-arch x64 \
  --linux "$kernel" \
  --initrd "$live_initramfs" \
  --uname "$kver" \
  --cmdline "$cmdline" \
  --os-release "@$rootfs/etc/os-release" \
  --stub "$rootfs/usr/lib/systemd/boot/efi/linuxx64.efi.stub" \
  --output "$builddir/osimage-live.efi"
if [ -n "${SECUREBOOT_PRIVATE_KEY:-}" ] && [ -n "${SECUREBOOT_CERTIFICATE:-}" ]; then
    set -- "$@" \
      --signtool systemd-sbsign \
      --secureboot-private-key "$SECUREBOOT_PRIVATE_KEY" \
      --secureboot-certificate "$SECUREBOOT_CERTIFICATE"
else
    echo '   no SECUREBOOT_PRIVATE_KEY/CERTIFICATE; building unsigned'
fi
ukify build "$@"
test -s "$builddir/osimage-live.efi"

echo '== GPT image with El Torito catalog (systemd-repart)'
cp -r "$common_dir/repart.d" "$builddir/repart.d"
esp_mib=$(( $(wc -c < "$builddir/osimage-live.efi") / 1024 / 1024 + 17 ))
mkdir -p "$builddir/repart.d/10-esp.conf.d"
printf '[Partition]\nSizeMinBytes=%sM\nSizeMaxBytes=%sM\n' "$esp_mib" "$esp_mib" \
  > "$builddir/repart.d/10-esp.conf.d/size.conf"

# Unprivileged builds map the build user to root so the squashfs is root-owned.
userns=
[ "$(id -u)" -eq 0 ] || userns="unshare --map-root-user"
# shellcheck disable=SC2086
TMPDIR="$(readlink -f "$builddir/tmp")" $userns systemd-repart \
  --offline=yes \
  --dry-run=no \
  --empty=create \
  --size=auto \
  --sector-size=512 \
  --seed=5b0c3ad2-7f43-4c51-9a1e-2d8e6f0b4c17 \
  --copy-source="$dist" \
  --definitions="$builddir/repart.d" \
  --el-torito=yes \
  --el-torito-volume=OSIMAGE_INSTALL \
  "$iso"
rm -rf "$builddir/tmp"

echo "built $iso ($(du -h "$iso" | cut -f1))"
