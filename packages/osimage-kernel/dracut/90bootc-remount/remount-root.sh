#!/bin/sh
# dracut mounts the root read-only by default; honour an explicit rw on
# the kernel command line so /etc and /var are writable before pivot.
if getargbool 0 rw; then
    mount -o remount,rw "$NEWROOT" >/dev/null 2>&1 || true
fi
