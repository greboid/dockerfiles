#!/bin/sh
test -s /usr/share/alpine-runner/storage.conf
grep -q '^mount_program = "/usr/bin/fuse-overlayfs"$' /usr/share/alpine-runner/storage.conf
# containers-common owns the default path; this package must never
# fight it for /etc/containers/storage.conf.
test ! -e /etc/containers/storage.conf
