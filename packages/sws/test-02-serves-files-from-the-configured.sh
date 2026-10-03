#!/bin/sh
mkdir -p /tmp/www
printf 'ok' > /tmp/www/index.html
SERVER_HOST=127.0.0.1 SERVER_PORT=8080 SERVER_ROOT=/tmp/www \
  /usr/bin/static-web-server >/tmp/sws.log 2>&1 &
pid=$!
sleep 1
code=$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:8080/ || true)
test "$code" = 200
kill "$pid"
