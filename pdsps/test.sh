#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/pdsps}"

podman run --rm "$IMAGE" -help 2>&1 | grep -q "pdsps" || true

