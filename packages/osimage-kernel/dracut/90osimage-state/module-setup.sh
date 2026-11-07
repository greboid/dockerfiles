#!/bin/bash
check() { return 0; }
depends() { return 0; }
install() {
    # After 90bootc-remount's composefs setup (pre-pivot 60), which
    # bind-mounts the persistent /var into $NEWROOT.
    inst_hook pre-pivot 70 "$moddir/etc-state.sh"
    inst_multiple mount mkdir cp chmod mv
    return 0
}
