#!/bin/sh
# The token database defaults to /data/db, absent in the guest and
# in the legacy image without a mounted /data.
if /usr/bin/irc-webhook >/tmp/irc-webhook.log 2>&1; then
  echo "expected missing-database failure" >&2
  exit 1
fi
grep -q "Unable to load config" /tmp/irc-webhook.log
grep -q "open /data/db" /tmp/irc-webhook.log
# With a writable token database the plugin helper's configuration
# check fires before any dial; nothing connects anywhere.
if DB_PATH=/tmp/tokens.db /usr/bin/irc-webhook >/tmp/irc-webhook2.log 2>&1; then
  echo "expected missing-token failure" >&2
  exit 1
fi
grep -q "Unable to create helper" /tmp/irc-webhook2.log
grep -q "plugin RPC token must be set" /tmp/irc-webhook2.log
