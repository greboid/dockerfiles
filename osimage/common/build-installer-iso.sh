#!/bin/sh
# Build a bootable installer ISO for the osimage image (UEFI-only, MBR).
set -eu

export SOURCE_DATE_EPOCH=0

dist="${1:-dist}"
rootfs="$dist/rootfs-installer"
test -d "$rootfs/usr/lib/modules" || { echo "missing $rootfs (run common/fetch-installer-rootfs.sh $dist, or build-installer-image.sh for a local build)" >&2; exit 1; }
test -s "$rootfs/usr/lib/systemd/boot/efi/linuxx64.efi.stub" || {
    echo "missing systemd-efistub in $rootfs" >&2; exit 1; }

command -v ukify >/dev/null || {
    echo 'ukify not found (install the org systemd ukify package)' >&2
    exit 1
}

ORG_REPO="${ORG_REPO:-https://git.mouse-lake.ts.net/api/packages/packages/alpine/v3.24/main}"
kver=$(ls "$rootfs/usr/lib/modules" | head -1)
kernel="$rootfs/usr/lib/modules/$kver/vmlinuz"
test -s "$kernel"

kernel_version=$(awk -F: '
    $1 == "P" { p = $2 }
    p == "osimage-kernel" && $1 == "V" { print $2; exit }
' "$rootfs/lib/apk/db/installed")
test -n "$kernel_version" || { echo "cannot read the osimage-kernel version from $rootfs/lib/apk/db/installed" >&2; exit 1; }

builddir="$dist/iso-installer"
builddir=$(readlink -f "$builddir")
iso="$dist/installer.iso"
rm -rf "$builddir" "$iso"
mkdir -p "$builddir/iso/LiveOS"

echo '== installer initramfs (osimage-kernel-initramfs-live package)'
rm -rf "$builddir/initramfs-live"
mkdir -p "$builddir/initramfs-live"
curl -fsSL --retry 3 "$ORG_REPO/x86_64/osimage-kernel-initramfs-live-$kernel_version.apk" \
  -o "$builddir/initramfs-live.apk"
tar -xzf "$builddir/initramfs-live.apk" -C "$builddir/initramfs-live"
live_initramfs="$builddir/initramfs-live/usr/lib/modules/$kver/initramfs-install.img"
test -s "$live_initramfs"

echo '== LiveOS squashfs (the single OS copy; uid forced to root because'
echo '   the extracted tree is owned by the build user)'
unset SOURCE_DATE_EPOCH
mksquashfs "$rootfs" "$builddir/iso/LiveOS/squashfs.img" \
  -comp zstd -Xcompression-level 15 -noappend -quiet -no-progress \
  -force-uid 0 -force-gid 0 -wildcards \
  -root-time 0 -root-uid 0 -root-gid 0 -mkfs-time 0 \
  -e 'proc/*' 'sys/*' 'dev/*' 'run/*' 'tmp/*' 'out/*' \
  -e 'boot/EFI/Linux/*.efi'
export SOURCE_DATE_EPOCH=0

echo "== live UKI (ukify; kernel + live initramfs + installer cmdline)"
cmdline="root=live:CDLABEL=OSIMAGE_INSTALL rw console=tty0 console=ttyS0,115200 init=/usr/lib/systemd/systemd"
signtool_args=
if [ -n "${SECUREBOOT_PRIVATE_KEY:-}" ] && [ -n "${SECUREBOOT_CERTIFICATE:-}" ]; then
    signtool_args="--signtool systemd-sbsign --secureboot-private-key $SECUREBOOT_PRIVATE_KEY --secureboot-certificate $SECUREBOOT_CERTIFICATE"
else
    echo '   no SECUREBOOT_PRIVATE_KEY/CERTIFICATE; building unsigned'
fi
# shellcheck disable=SC2086
ukify build \
  --efi-arch x64 \
  --linux "$kernel" \
  --initrd "$live_initramfs" \
  --uname "$kver" \
  --cmdline "$cmdline" \
  --os-release "@$rootfs/etc/os-release" \
  --stub "$rootfs/usr/lib/systemd/boot/efi/linuxx64.efi.stub" \
  $signtool_args \
  --output "$builddir/osimage-live.efi"
test -s "$builddir/osimage-live.efi"

echo '== ESP FAT image (El Torito payload and the USB MBR type 0xEF partition;'
echo '   BOOTX64.EFI runs straight from the firmware, no second boot manager)'
payload=$(wc -c < "$builddir/osimage-live.efi")
fat_mib=$(( payload / 1024 / 1024 + 16 + 1 ))
dd if=/dev/zero of="$builddir/efi.img" bs=1M count="$fat_mib" status=none
# mtools takes timestamps from SOURCE_DATE_EPOCH (in local time); FAT cannot store dates before 1980.
fat_env="env TZ=UTC SOURCE_DATE_EPOCH=315532800"
$fat_env mformat -F -N 00000000 -i "$builddir/efi.img" -v OSIMAGEEFI ::
$fat_env mmd -i "$builddir/efi.img" ::/EFI ::/EFI/BOOT
$fat_env mcopy -i "$builddir/efi.img" "$builddir/osimage-live.efi" ::/EFI/BOOT/BOOTX64.EFI
mdir -i "$builddir/efi.img" ::/EFI/BOOT | grep -q BOOTX64

echo '== BIOS stub (x86 El Torito entry; forces the 0xEF section header layout)'
python3 - "$builddir/iso/boot/bios-stub.bin" <<'PYEOF'
import os
import sys

os.makedirs(os.path.dirname(sys.argv[1]), exist_ok=True)
msg = b"osimage installer: BIOS boot is not supported; boot this media in UEFI mode.\r\n"
code = bytes([
    0xBE, 0x13, 0x00,   # mov si, msg
    0xAC,               # lodsb
    0x08, 0xC0,         # or al, al
    0x74, 0x08,         # jz halt
    0xB4, 0x0E,         # mov ah, 0x0e
    0xB3, 0x00,         # mov bh, 0x00
    0xCD, 0x10,         # int 0x10
    0xEB, 0xF3,         # jmp lodsb
    0xF4,               # halt: hlt
    0xEB, 0xFD,         # jmp halt
])
img = bytearray(code) + msg + b'\x00'
img += b'\xF4' * (2048 - len(img))
open(sys.argv[1], 'wb').write(bytes(img))
PYEOF

echo '== assemble hybrid MBR + El Torito ISO'
cat > "$builddir/iso/README.txt" <<EOF
osimage installer (UEFI boot; CD-ROM or dd to USB).

Attached as a CD-ROM (or burned to optical media), the firmware boots
the El Torito EFI entry: a FAT image whose BOOTX64.EFI is a UKI
(systemd-stub) holding the kernel and the installer initramfs with the
live command line embedded. Written to a USB stick, the firmware sees
the same FAT image as an MBR type 0xEF partition and boots the same
UKI from it. Either way the dmsquash-live initramfs mounts the
LiveOS squashfs on this ISO filesystem — the standard osimage build as
the live installer root.
osimage-install pulls the published image from the package registry
(anonymously pullable; set REGISTRY_AUTH_FILE for private targets) and
installs it with systemd-repart layout (bootc install to-disk
--run-repart), so the installed digest is the registry digest and the
upgrade origin is the tag itself. Network access is required at
install time; override the refs with OSIMAGE_SOURCE_REF /
OSIMAGE_TARGET_REF.

Boot the stick, log in as root, install with:

  osimage-install /dev/<disk>
EOF
find "$builddir/iso" -exec touch -h -d @0 {} +
xorriso -as mkisofs \
  -o "$iso" \
  -r -V OSIMAGE_INSTALL \
  -append_partition 2 C12A7328-F81F-11D2-BA4B-00A0C93EC93B "$builddir/efi.img" \
  -eltorito-platform 0 -b boot/bios-stub.bin --no-emul-boot -boot-load-size 4 \
  -eltorito-alt-boot -e --interval:appended_partition_2:all:: --no-emul-boot \
  "$builddir/iso"

# xorriso leaves Sector Count 0, which makes EDK2 drop an entry pointing past the ISO volume end.
python3 - "$iso" "$builddir/efi.img" <<'PYEOF'
import struct
import sys

iso, fat = sys.argv[1], sys.argv[2]
fat_bytes = len(open(fat, 'rb').read())
sector_count = fat_bytes // 2048
assert sector_count <= 0xFFFF, f'El Torito Sector Count overflow ({sector_count}); revert to in-tree efi.img'
with open(iso, 'r+b') as f:
    f.seek(33 * 2048 + 96 + 6)  # catalog LBA 33, EFI section entry, Sector Count
    f.write(struct.pack('<H', sector_count))
print(f'catalog EFI entry Sector Count set to {sector_count} ({fat_bytes} bytes)')
PYEOF

echo "built $iso ($(du -h "$iso" | cut -f1))"
