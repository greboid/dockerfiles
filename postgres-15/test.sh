#!/bin/sh
set -e


IMAGE="${1:-reg.g5d.dev/postgres-15}"
PORT=15432

CID=$(podman run -d --name test-postgres -p $PORT:5432 \
  -e POSTGRES_PASSWORD=test123 "$IMAGE" 2>/dev/null)

trap "podman rm -f $CID >/dev/null 2>&1 || true" EXIT

sleep 10

IP=$(podman inspect -f '{{.NetworkSettings.IPAddress}}' $CID 2>/dev/null)

podman run --rm --network host \
  alpine:latest \
  sh -c "apk add --no-cache postgresql-client > /dev/null 2>&1 && pg_isready -h localhost -p $PORT" > /dev/null 2>&1 || \
  (podman rm -f $CID >/dev/null 2>&1; exit 1)

podman run --rm --network host \
  alpine:latest \
  sh -c "apk add --no-cache postgresql-client > /dev/null 2>&1 && PGPASSWORD=test123 psql -h localhost -p $PORT -U postgres -c 'SELECT 1;'" | grep -q 1 || \
  (podman rm -f $CID >/dev/null 2>&1; exit 1)

