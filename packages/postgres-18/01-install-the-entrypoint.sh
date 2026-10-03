#!/bin/sh
# The legacy image's entrypoint, verbatim; installed into the package.
install -Dm 0755 entry.sh "${DESTDIR}/usr/bin/entry.sh"
