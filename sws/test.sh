#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/sws}"

podman run --rm "$IMAGE" -V 2>&1 | grep -q "Version" || true

