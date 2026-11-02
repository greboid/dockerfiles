#!/bin/bash
check() { return 0; }
depends() { echo fs-lib; }
installkernel() { return 0; }
install() {
    inst_hook cmdline 95 "$moddir/parse-gptroot.sh"
    inst_hook mount 98 "$moddir/gptroot-mount.sh"
    inst_multiple lsblk mount
}
