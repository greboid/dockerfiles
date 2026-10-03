#!/bin/sh
# Fully offline: loopback bind, local storage, embedded templates.
mkdir -p /tmp/linx-files /tmp/linx-meta
/usr/bin/linx-server -bind=127.0.0.1:8080 -filespath=/tmp/linx-files/ \
  -metapath=/tmp/linx-meta/ -sitename=linx -allowhotlink \
  >/tmp/linx-server.log 2>&1 &
pid=$!
sleep 1
code=$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:8080/ || true)
test "$code" = 200
kill "$pid"
