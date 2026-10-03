#!/bin/sh
install -Dm755 src/redis-server "${DESTDIR}/usr/bin/redis"
# Legacy VOLUME: persistence for the configured database directory.
printf '%s\n' \
  'bind 0.0.0.0' \
  'protected-mode no' \
  '' \
  'dir /home/nonroot/database' \
  '' \
  'port 6379' \
  > "${DESTDIR}/redis.conf"
