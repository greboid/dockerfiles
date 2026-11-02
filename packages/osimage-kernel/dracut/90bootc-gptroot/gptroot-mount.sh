#!/bin/sh
# root=gpt-auto: resolve systemd's discoverable GPT root partition and
# mount it at $NEWROOT. The busybox dracut initramfs has no
# systemd-gpt-auto-generator, and rootfs-block only handles roots with
# a block: prefix, so a sealed UKI (whose cmdline is baked at build
# time and cannot carry a per-install root=UUID) relies on this.
command -v getarg > /dev/null || . /lib/dracut-lib.sh

[ "$(getarg root=)" = "gpt-auto" ] || return 0

if ismounted "$NEWROOT"; then
    return 0
fi

# Linux root partitions for the known architectures (x86-64, arm64,
# riscv32/64); we ship x86-64 media but scanning the list costs nothing.
for guid in \
    4f68bce3-e8cd-4db1-96e7-fbcaf984b709 \
    b921b045-1df0-41c3-af44-4c6f280d3fae \
    60d5a7fe-f29b-484c-b1ef-8a5d2ed48e33 \
    72ec70a6-cf74-40e6-bd49-4bda08e8f224; do
    # udev may still be enumerating; nothing else registered a root
    # device to wait for, so poll briefly.
    for i in 1 2 3 4 5 6 7 8 9 10; do
        rootdev=$(lsblk -rno PATH,PARTTYPE 2>/dev/null |
            awk -v g="$guid" 'tolower($2) == g { print $1; exit }')
        [ -n "$rootdev" ] && break
        sleep 1
    done
    [ -n "$rootdev" ] && break
done

if [ -z "$rootdev" ]; then
    warn "bootc-gptroot: no discoverable root partition found"
    return 1
fi

mkdir -p "$NEWROOT"
# busybox mount's fs auto-probing fails with ENOENT here; be explicit.
info "bootc-gptroot: mounting discoverable root $rootdev"
mount -t ext4 -o ro "$rootdev" "$NEWROOT"
