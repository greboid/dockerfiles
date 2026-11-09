#!/bin/bash
check() { return 0; }
depends() { return 0; }
install() {
    inst_hook pre-pivot 70 "$moddir/etc-state.sh"
    inst_multiple mount mkdir cp chmod mv rm
    return 0
}
