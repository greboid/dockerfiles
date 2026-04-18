#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/identd}"

podman run --rm "$IMAGE" -h 2>&1 | grep -q "identd\|port\|response"

