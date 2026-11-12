#!/bin/sh
# root=gpt-auto: mount the discoverable GPT root partition at $NEWROOT
# (this initramfs has no systemd-gpt-auto-generator).
command -v getarg > /dev/null || . /lib/dracut-lib.sh

[ "$(getarg root=)" = "gpt-auto" ] || return 0

if ismounted "$NEWROOT"; then
    return 0
fi

ROOT_GUIDS="4f68bce3-e8cd-4db1-96e7-fbcaf984b709 b921b045-1df0-41c3-af44-4c6f280d3fae 60d5a7fe-f29b-484c-b1ef-8a5d2ed48e33 72ec70a6-cf74-40e6-bd49-4bda08e8f224"
ESP_GUID=c12a7328-f81f-11d2-ba4b-00a0c93ec93b

disk_of() {
    pk=$(lsblk -nrno PKNAME "$1" 2>/dev/null)
    [ -n "$pk" ] && {
        echo "$pk"
        return 0
    }
    echo "${1##*/}"
}

root_parts() {
    _scope=${1:-}
    [ -n "$_scope" ] && _scope="/dev/$_scope"
    for guid in $ROOT_GUIDS; do
        # shellcheck disable=SC2086
        lsblk -rno PATH,PARTTYPE $_scope 2>/dev/null |
            awk -v g="$guid" 'tolower($2) == g { print $1 }'
    done
}

for _ in 1 2 3 4 5 6 7 8 9 10; do
    [ -n "$(root_parts)" ] && break
    sleep 1
done

bootdisk=
if [ ! -r /sys/firmware/efi/efivars ]; then
    mount -t efivarfs efivarfs /sys/firmware/efi/efivars 2>/dev/null || true
fi
espuuid=$(dd if=/sys/firmware/efi/efivars/LoaderDevicePartUUID-4a67b082-0a4c-41cf-b6c7-440b29bb8c4f bs=4 skip=1 2>/dev/null |
    tr -d '\0' | tr 'A-F' 'a-f')
if [ -n "$espuuid" ]; then
    espdev=$(lsblk -rno PATH,PARTUUID,PARTTYPE 2>/dev/null |
        awk -v g="$ESP_GUID" -v u="$espuuid" 'tolower($2) == u && tolower($3) == g { print $1; exit }')
    [ -n "$espdev" ] && bootdisk=$(disk_of "$espdev")
    [ -n "$bootdisk" ] ||
        warn "bootc-gptroot: loader ESP ($espuuid) not found; falling back to all-disk scan"
fi

if [ -n "$bootdisk" ]; then
    if [ "$(root_parts "$bootdisk" | wc -l)" -eq 1 ]; then
        rootdev=$(root_parts "$bootdisk")
    else
        warn "bootc-gptroot: expected exactly one discoverable root on the booted disk $bootdisk"
    fi
else
    allroots=$(root_parts)
    case $(printf '%s\n' "$allroots" | grep -c .) in
        1) rootdev=$allroots ;;
        0) : ;;
        *)
            warn "bootc-gptroot: discoverable root partitions on multiple disks; refusing to guess"
            return 1
            ;;
    esac
fi

if [ -z "$rootdev" ]; then
    if [ -n "$bootdisk" ]; then
        warn "bootc-gptroot: no discoverable root partition on the booted disk $bootdisk"
    else
        warn "bootc-gptroot: no discoverable root partition found"
    fi
    return 1
fi

mkdir -p "$NEWROOT"
info "bootc-gptroot: mounting discoverable root $rootdev"
mount -t ext4 -o rw "$rootdev" "$NEWROOT"
