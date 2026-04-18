#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/goplum}"

podman run --rm "$IMAGE" -help 2>&1 | grep -q "goplum" || true

