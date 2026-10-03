#!/bin/sh
# End-to-end HEIC check: a real HEVC-branded HEIC (heif-enc via the
# x265 plugin) decoded with the shipped sharp through
# libvips -> libheif -> libde265.
ffmpeg -loglevel error -f lavfi -i color=c=black:s=64x64 -frames:v 1 -y /tmp/smoke.png
heif-enc -q 60 -o /tmp/smoke.heic /tmp/smoke.png
node -e 'require("/usr/src/app/server/node_modules/sharp")("/tmp/smoke.heic").resize(32).jpeg().toBuffer().then((b) => { if (b.length === 0) { throw new Error("empty jpeg"); } console.log("HEIC smoke test ok"); });' | grep -q "HEIC smoke test ok"
