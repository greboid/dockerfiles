#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/irc-webhook}"

podman run --rm "$IMAGE" -h 2>&1 | grep -q "irc-webhook\|admin-key\|channel"
