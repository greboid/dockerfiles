#!/bin/sh
# A missing -config file aborts before any listener or IRC network
# is touched.
if /usr/bin/soju -config /nonexistent.cfg >/tmp/soju.log 2>&1; then
  echo "expected missing-config failure" >&2
  exit 1
fi
grep -q "failed to load config file" /tmp/soju.log
