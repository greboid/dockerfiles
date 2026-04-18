#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/cocoon}"

podman run --rm "$IMAGE" -v >/dev/null 2>&1 || true

