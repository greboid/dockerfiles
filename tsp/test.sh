#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/tsp}"

podman run --rm "$IMAGE" -h 2>&1 | grep -q "tsp\|tailscale-auth-key\|upstream"
