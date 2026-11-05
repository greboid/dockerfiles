#!/bin/bash
check() { return 0; }
depends() { return 0; }
install() {
    inst_hook pre-pivot 50 "$moddir/remount-root.sh"
    inst_hook pre-pivot 60 "$moddir/composefs-setup.sh"
    inst_multiple mount
    inst_multiple /usr/lib/bootc/initramfs-setup
    # initramfs-setup reads setup-root-conf.toml from the initramfs
    # filesystem (missing file means: defaults, which bind-mount /etc
    # read-write). Mirror the bootc module's embedding so this module is
    # self-sufficient; skip silently when the image ships no config.
    if [ -e /usr/lib/composefs/setup-root-conf.toml ]; then
        inst_simple /usr/lib/composefs/setup-root-conf.toml
    fi
    return 0
}
