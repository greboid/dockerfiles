#!/bin/sh
set -e

IMAGE="${1:-reg.g5d.dev/typst}"

podman run --rm "$IMAGE" --version 2>&1 | grep -q "Typst" || true
