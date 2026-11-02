#!/bin/bash
check() { return 0; }
depends() { return 0; }
install() {
    inst_hook pre-pivot 50 "$moddir/remount-root.sh"
    inst_hook pre-pivot 60 "$moddir/composefs-setup.sh"
    inst_multiple mount
    inst_multiple /usr/lib/bootc/initramfs-setup
}
