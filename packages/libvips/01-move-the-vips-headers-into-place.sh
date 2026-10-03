#!/bin/sh
# sharp's <vips/vips8> needs the headers in the vips/ namespace, but
# meson installs them flat (with version.h under vips/). Move the
# flat headers in; drop the two internal headers whose names collide
# with musl-dev (nothing in the public include chain uses them — the
# legacy dodged this via the isolated /usr/local prefix).
mkdir -p "${DESTDIR}/usr/include/vips"
find "${DESTDIR}/usr/include" -maxdepth 1 -name '*.h' \
  -exec mv {} "${DESTDIR}/usr/include/vips/" \;
if [ -f "${DESTDIR}/usr/include/vips8" ]; then
  mv "${DESTDIR}/usr/include/vips8" \
     "${DESTDIR}/usr/include/vips/vips8"
fi
rm -f "${DESTDIR}/usr/include/memory.h" \
      "${DESTDIR}/usr/include/semaphore.h"
find "${DESTDIR}" -name "*.la" -delete
