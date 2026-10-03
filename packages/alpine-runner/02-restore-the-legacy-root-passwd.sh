#!/bin/sh
# The legacy minirootfs parent had root in /etc/passwd; buildah
# resolves the current user through it ("unknown userid 0" without
# this). apko's accounts gizmo appends its declared users to these
# package-provided files (nonroot 65532 comes from the image).
install -d -m 0755 "${DESTDIR}/etc"
printf '%s\n' 'root:x:0:0:root:/root:/bin/sh' > "${DESTDIR}/etc/passwd"
printf '%s\n' 'root:x:0:' > "${DESTDIR}/etc/group"
