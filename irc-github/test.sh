#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/irc-github}"

podman run --rm "$IMAGE" -h 2>&1 | grep -q "irc-github\|channel\|github-secret"
