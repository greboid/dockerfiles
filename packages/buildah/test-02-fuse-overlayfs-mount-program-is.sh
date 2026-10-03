#!/bin/sh
# Dependency-closure check only. Melange guests cannot grant what
# exercising containers for real needs (user namespaces, /dev/fuse,
# mount capability), so probes of that shape end up testing the
# guest kernel rather than this package — and fail differently on
# every runner. Real container behaviour is verified on a
# privileged host by tooling/test-build-images.sh (buildah
# bud/from/mount smoke).
test -x /usr/bin/fuse-overlayfs
fuse-overlayfs --version
