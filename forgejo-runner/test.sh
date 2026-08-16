#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/forgejo-runner}"

podman run --rm "$IMAGE" --version 2>&1 | grep -q forgejo-runner
podman run --rm --entrypoint /usr/bin/git "$IMAGE" --version | grep -q "git version"
podman run --rm --entrypoint /bin/bash "$IMAGE" --version | grep -q "GNU bash"
