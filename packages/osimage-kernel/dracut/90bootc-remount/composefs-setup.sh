#!/bin/sh
# Run bootc's composefs root setup (busybox-initramfs equivalent of
# bootc-root-setup.service).
#
# A supplied deployment digest that cannot be set up must fail the
# boot: continuing on the plain root would silently run whatever the
# filesystem happens to hold, bypassing the composefs deployment the
# UKI names (this is what rejects a sealed tree modified after the UKI
# was built). Boots without a digest argument (factory image, live
# media, direct-kernel diagnostics) intentionally keep the plain root.
command -v getarg > /dev/null || . /lib/dracut-lib.sh

if [ -n "$(getarg composefs.digest=)" ] || [ -n "$(getarg composefs=)" ]; then
    info "bootc: setting up composefs root"
    if ! /usr/lib/bootc/initramfs-setup --sysroot "$NEWROOT"; then
        warn "bootc: composefs setup failed for the supplied deployment digest"
        # A pre-pivot hook failure alone does not stop dracut; the pivot
        # would continue on the plain root. Drop to the emergency shell
        # so the boot cannot proceed without the named deployment.
        emergency
        return 1
    fi
fi
