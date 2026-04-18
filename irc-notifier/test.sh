#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/irc-notifier}"

podman run --rm "$IMAGE" -h 2>&1 | grep -q "irc-notifier\|igloo-token\|network"
