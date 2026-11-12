#!/bin/sh
# Run bootc's composefs root setup (busybox-initramfs equivalent of
# bootc-root-setup.service).
command -v getarg > /dev/null || . /lib/dracut-lib.sh

if [ -n "$(getarg composefs.digest=)" ] || [ -n "$(getarg composefs=)" ]; then
    info "bootc: setting up composefs root"
    if ! /usr/lib/bootc/initramfs-setup --sysroot "$NEWROOT"; then
        warn "bootc: composefs setup failed for the supplied deployment digest"
        emergency
        return 1
    fi
fi
