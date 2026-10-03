#!/bin/sh
# Fails before any connection attempt; no IRC network is touched.
if /usr/bin/irc-bot >/tmp/irc-bot.log 2>&1; then
  echo "expected missing-server failure" >&2
  exit 1
fi
grep -q "Server is mandatory" /tmp/irc-bot.log
