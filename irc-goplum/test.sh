#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/irc-goplum}"

podman run --rm "$IMAGE" -h 2>&1 | grep -q "irc-goplum\|channel\|secret"
