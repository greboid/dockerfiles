#!/bin/sh
# The legacy build pruned vendored frontend notice trees.
rm -rf "${DESTDIR}/notices/forgejo.org/node_modules" \
       "${DESTDIR}/notices/forgejo.org/public" \
       "${DESTDIR}/notices/forgejo.org/forgejo"
