#!/bin/sh
# Fails before the tailscale client starts; no network is touched.
if /usr/bin/thp >/tmp/thp.log 2>&1; then
  echo "expected missing-upstream failure" >&2
  exit 1
fi
grep -q "Upstream host cannot be blank" /tmp/thp.log
