#!/bin/sh
# The plugin helper validates configuration before any dial; nothing
# connects anywhere.
if /usr/bin/irc-goplum >/tmp/irc-goplum.log 2>&1; then
  echo "expected missing-token failure" >&2
  exit 1
fi
grep -q "Unable to create plugin helper" /tmp/irc-goplum.log
grep -q "plugin RPC token must be set" /tmp/irc-goplum.log
