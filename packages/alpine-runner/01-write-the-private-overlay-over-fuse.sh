#!/bin/sh
install -d -m 0755 "${DESTDIR}/usr/share/alpine-runner"
# Verbatim legacy storage.conf (containers/alpine-runner/storage.conf):
# overlay driver with fuse-overlayfs as the unprivileged mount program.
printf '%s\n' \
  '[storage]' \
  'driver = "overlay"' \
  '' \
  '[storage.options.overlay]' \
  'mount_program = "/usr/bin/fuse-overlayfs"' \
  > "${DESTDIR}/usr/share/alpine-runner/storage.conf"
chmod 0644 "${DESTDIR}/usr/share/alpine-runner/storage.conf"
