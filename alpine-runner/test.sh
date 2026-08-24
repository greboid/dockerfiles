#!/bin/sh
set -e
IMAGE="${1:-reg.g5d.dev/alpine-runner}"
PODMAN="${PODMAN:-podman}"

CID=$($PODMAN run -d --entrypoint /bin/sh "$IMAGE" -c 'sleep 60')
trap '$PODMAN rm -f $CID >/dev/null 2>&1 || true' EXIT

sleep 2

$PODMAN exec $CID buildah --version
$PODMAN exec $CID buildah --help >/dev/null
$PODMAN exec $CID buildah from --name testctr alpine:3.22 >/dev/null
echo "buildah is functional"
