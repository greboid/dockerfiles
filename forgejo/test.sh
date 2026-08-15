#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/forgejo}"
PORT=13000

CID=$(podman run -d -p $PORT:3000 "$IMAGE")

trap "podman rm -f $CID >/dev/null 2>&1 || true" EXIT

for i in $(seq 1 30); do
  if curl -f -s http://localhost:$PORT/ > /dev/null 2>&1; then
    break
  fi
  sleep 1
done

curl -f -s http://localhost:$PORT/ | grep -qi forgejo
