#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/redis}"

podman run --rm "$IMAGE" -h 2>&1 | grep -q "redis-server\|daemonize"

