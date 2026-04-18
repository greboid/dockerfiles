#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/webhooked}"

podman run --rm "$IMAGE" -h 2>&1 | grep -q "webhooked\|url\|token"
