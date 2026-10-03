#!/bin/sh
mkdir -p /tmp/goplum && cd /tmp/goplum
openssl req -x509 -newkey rsa:2048 -nodes -keyout ca.key -out ca.crt \
  -days 1 -subj /CN=smoke-ca >/dev/null 2>&1
openssl req -newkey rsa:2048 -nodes -keyout goplum.key -out goplum.csr \
  -subj /CN=goplum >/dev/null 2>&1
openssl x509 -req -in goplum.csr -CA ca.crt -CAkey ca.key \
  -out goplum.crt -days 1 >/dev/null 2>&1
printf 'check debug.random "one" {\n    interval = 2s\n}\nalert debug.sysout "text" {}\n' \
  > goplum.conf
/usr/bin/goplum -config goplum.conf -cert goplum.crt -key goplum.key -ca-cert ca.crt \
  >goplum.log 2>&1 &
pid=$!
sleep 3
grep -q "Starting API server on port 7586" goplum.log
grep -q "Check 'one' executed" goplum.log
nc -z 127.0.0.1 7586
kill "$pid"
