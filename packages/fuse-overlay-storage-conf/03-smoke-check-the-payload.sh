#!/bin/sh
test -s "${DESTDIR}/usr/share/alpine-runner/storage.conf"
grep -q '^driver = "overlay"$' "${DESTDIR}/usr/share/alpine-runner/storage.conf"
grep -q '^mount_program = "/usr/bin/fuse-overlayfs"$' "${DESTDIR}/usr/share/alpine-runner/storage.conf"
grep -q '^root:x:0:0:root:/root:/bin/sh$' "${DESTDIR}/etc/passwd"
