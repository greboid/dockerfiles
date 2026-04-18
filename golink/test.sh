#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/golink}"

podman run --rm "$IMAGE" -h 2>&1 | grep -q "golink\|sqlitedb"
