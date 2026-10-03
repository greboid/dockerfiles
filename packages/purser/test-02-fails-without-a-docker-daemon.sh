#!/bin/sh
# The initial scan lists containers first, so it fails before any
# vulnerability database is downloaded.
if /usr/bin/purser >/tmp/purser.log 2>&1; then
  echo "expected missing-daemon failure" >&2
  exit 1
fi
grep -q "Failed to scan containers" /tmp/purser.log
