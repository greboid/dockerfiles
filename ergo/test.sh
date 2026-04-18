#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/ergo}"

podman run --rm "$IMAGE" -h 2>&1 | grep -q "ergo\|initdb\|run"
