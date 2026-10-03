#!/bin/sh
# Fails before the tailscale client starts; no network is touched.
if /usr/bin/tsp >/tmp/tsp.log 2>&1; then
  echo "expected missing-upstream failure" >&2
  exit 1
fi
grep -q "Error parsing upstream address" /tmp/tsp.log
