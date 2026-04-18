#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/httpredirect}"
PORT=18080

CID=$(podman run -d --name test-httpredirect -p $PORT:8080 "$IMAGE" 2>/dev/null)

trap "podman rm -f $CID >/dev/null 2>&1 || true" EXIT

for i in $(seq 1 30); do
  if curl -f -s http://localhost:$PORT/ > /dev/null 2>&1; then
    break
  fi
  sleep 1
done

curl -f -s http://localhost:$PORT/ > /dev/null

