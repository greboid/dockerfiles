#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/dind}"

CID=$(podman run --privileged -d "$IMAGE")

trap "podman rm -f $CID >/dev/null 2>&1 || true" EXIT

for i in $(seq 1 60); do
  if podman exec $CID docker version >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

podman exec $CID docker version --format '{{.Server.Version}}' | grep -q .
podman exec $CID docker run --rm hello-world | grep -q "Hello from Docker!"
