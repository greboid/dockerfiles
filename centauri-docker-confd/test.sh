#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/centauri-docker-confd}"

podman run --rm "$IMAGE" -h 2>&1 | grep -q "centauri-docker-confd\|listen\|proxytag"
