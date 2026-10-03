#!/bin/sh
# buildah's containers-common owns /etc/containers/storage.conf, so
# the legacy overlay-over-fuse configuration ships at a private path
# and the image points CONTAINERS_STORAGE_CONF at it.
printf '%s\n' \
  '[storage]' \
  'driver = "overlay"' \
  '' \
  '[storage.options.overlay]' \
  'mount_program = "/usr/bin/fuse-overlayfs"' \
  > "${DESTDIR}/usr/share/forgejo-runner/storage.conf"
