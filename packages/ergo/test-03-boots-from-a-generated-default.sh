#!/bin/sh
mkdir /tmp/ircd && cd /tmp/ircd
/usr/bin/ergo defaultconfig > ircd.yaml
test -s ircd.yaml
grep -q 'network' ircd.yaml
/usr/bin/ergo mkcerts --conf ircd.yaml >/dev/null
test -s fullchain.pem && test -s privkey.pem
# Bundled translations live at /ircd-bin/languages (legacy path).
ln -s /ircd-bin/languages languages
# --smoke starts the server, verifies it, and exits (nonzero on failure).
/usr/bin/ergo run --conf ircd.yaml --smoke >/tmp/ergo-run.log 2>&1
grep -q 'Server running' /tmp/ergo-run.log
