#!/bin/sh
# containuum connects to the Docker socket, absent in the test guest.
if /usr/bin/centauri-docker-confd >/tmp/confd.log 2>&1; then
  echo "expected failure without a Docker daemon" >&2
  exit 1
fi
grep -qiE "docker|containuum|failed" /tmp/confd.log
