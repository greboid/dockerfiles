#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/irc-distribution}"

podman run --rm "$IMAGE" -h 2>&1 | grep -q "irc-distribution\|channel\|rpc-host"
