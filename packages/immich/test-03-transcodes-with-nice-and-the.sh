#!/bin/sh
# Immich hardcodes fluent-ffmpeg niceness 10 (nice must resolve) and
# HDR thumbnails require jellyfin's tonemapx filter — the two
# runtime quirks that motivated the legacy image's tooling.
nice -n 10 true
ffmpeg -loglevel error -f lavfi -i testsrc2=s=256x144:r=5 -t 1 -pix_fmt p010le -c:v libx265 -x265-params lossless=1 -color_primaries bt2020 -color_trc smpte2084 -colorspace bt2020nc -y /tmp/smoke-hdr.mp4
ffmpeg -loglevel error -i /tmp/smoke-hdr.mp4 -vf 'tonemapx=tonemap=hable:desat=0:p=bt709:t=bt709:m=bt709:r=pc:peak=100:format=yuv420p' -frames:v 1 -y /tmp/smoke-tonemap.jpg
test -s /tmp/smoke-tonemap.jpg
