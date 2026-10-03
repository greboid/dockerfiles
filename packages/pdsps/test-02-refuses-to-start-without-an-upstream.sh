#!/bin/sh
if /usr/bin/pdsps >/tmp/pdsps.log 2>&1; then
  echo "expected missing-upstream failure" >&2
  exit 1
fi
grep -q "upstream is required" /tmp/pdsps.log
