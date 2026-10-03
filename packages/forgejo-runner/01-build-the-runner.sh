#!/bin/sh
# The Makefile stamps ver.version=v<VERSION> and adds
# -extldflags "-static" netgo osusergo itself.
make build VERSION=${PACKAGE_VERSION}
install -Dm755 forgejo-runner "${DESTDIR}/usr/bin/forgejo-runner"
