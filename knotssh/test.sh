#!/bin/sh
set -e
IMAGE="${1:-reg.g5d.dev/knot-ssh}"
podman run --rm --entrypoint /usr/bin/git "$IMAGE" --version
