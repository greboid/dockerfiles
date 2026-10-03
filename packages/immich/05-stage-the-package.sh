#!/bin/sh
mkdir -p "${DESTDIR}/usr/src/app" "${DESTDIR}/usr/bin"
cp -a /home/build/out/server "${DESTDIR}/usr/src/app/server"
cp -a web/build "${DESTDIR}/usr/src/app/www"
cp -a geodata "${DESTDIR}/usr/src/app/geodata"
install -d -m 0755 "${DESTDIR}/usr/src/app/upload"
# jellyfin-ffmpeg ships no PATH entries; Immich resolves ffmpeg and
# ffprobe through PATH.
ln -sf ../lib/jellyfin-ffmpeg/ffmpeg "${DESTDIR}/usr/bin/ffmpeg"
ln -sf ../lib/jellyfin-ffmpeg/ffprobe "${DESTDIR}/usr/bin/ffprobe"
