#!/bin/sh
# Flag validation fails before any tailnet or WireGuard contact.
if /usr/bin/tsv --tailscale-config-dir=/config >/tmp/tsv.log 2>&1; then
  echo "expected missing-configuration failure" >&2
  exit 1
fi
grep -q "Flag validation failed" /tmp/tsv.log
grep -q "\-\-wg-private-key is required" /tmp/tsv.log
