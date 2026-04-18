#!/bin/sh

echo "Cleaning up test containers..."
podman ps -a --format "{{.Names}}" 2>/dev/null | grep "^test-" | xargs -r podman rm -f

echo "Cleaning up temporary directories..."
rm -rf /tmp/golink-test /tmp/redis-test /tmp/linx-files /tmp/linx-meta \
       /tmp/purser-cache /tmp/purser-output /tmp/thp-config \
       /tmp/registryauth-data /tmp/centauri-data 2>/dev/null || true

echo "Done!"
