#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/linx-server}"
PORT=18080

mkdir -p /tmp/linx-files /tmp/linx-meta
CID=$(podman run -d --name test-linx -p $PORT:8080 \
  -v /tmp/linx-files:/data/files \
  -v /tmp/linx-meta:/data/meta \
  "$IMAGE" 2>/dev/null)

trap "podman rm -f $CID >/dev/null 2>&1 || true; rm -rf /tmp/linx-files /tmp/linx-meta" EXIT

for i in $(seq 1 30); do
  if curl -f -s http://localhost:$PORT/ > /dev/null 2>&1; then
    break
  fi
  sleep 1
done

curl -f -s http://localhost:$PORT/ > /dev/null

curl -f -s -o /dev/null -w "%{http_code}" http://localhost:$PORT/ | grep -q "200\|400\|405"

