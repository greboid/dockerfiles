#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/dockercleanup}"

podman run --rm "$IMAGE" -help 2>&1 | grep -q "dockercleanup" || true

