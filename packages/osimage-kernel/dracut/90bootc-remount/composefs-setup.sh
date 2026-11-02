#!/bin/sh
# Run bootc's composefs root setup (busybox-initramfs equivalent of
# bootc-root-setup.service). When no deployment digest matches - e.g. a
# factory image booted before bootc install - fall back to the plain root.
if [ -n "$(getarg composefs.digest=)" ] || [ -n "$(getarg composefs=)" ]; then
    info "bootc: setting up composefs root"
    if ! /usr/lib/bootc/initramfs-setup --sysroot "$NEWROOT"; then
        warn "bootc: composefs setup failed; continuing with plain root"
        # initramfs-setup may leave an empty mount stacked on the sysroot;
        # recover the underlying root if so.
        if [ ! -e "$NEWROOT/sbin/init" ] && [ ! -e "$NEWROOT/usr/lib/systemd/systemd" ]; then
            info "bootc: sysroot lacks init; attempting to recover underlying root"
            umount "$NEWROOT" 2>/dev/null
            mount | grep -E "sysroot|vda " | while read -r l; do info "mount: $l"; done
            ls "$NEWROOT" | head -5 | while read -r l; do info "sysroot entry: $l"; done
        fi
    fi
fi
