#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/tailscale}"
PODMAN="${PODMAN:-podman}"

$PODMAN run --rm --entrypoint /usr/local/bin/tailscale "$IMAGE" version | grep -qE '^[0-9]+\.[0-9]+'

CID=$($PODMAN run -d "$IMAGE")
trap '$PODMAN rm -f -t 1 "$CID" >/dev/null 2>&1 || true' EXIT
sleep 8
$PODMAN exec "$CID" tailscale status 2>&1 | grep -q "Logged out."
