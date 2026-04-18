#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/irc-bot}"

podman run --rm "$IMAGE" -help 2>&1 | grep -q "irc-bot\|bot" || true

