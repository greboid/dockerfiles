#!/bin/sh
PORT=9000 RESPONSE=smoke /usr/bin/identd >/tmp/identd.log 2>&1 &
pid=$!
sleep 1
# <port-on-server>, <port-on-client> queries return USERID lines.
reply=$(printf '6115, 6667\r\n' | nc -w 3 127.0.0.1 9000)
echo "$reply" | grep -q '6115, 6667 : USERID : UNIX : smoke'
kill "$pid"
# Defaults: port 8000, response "ident".
/usr/bin/identd >/tmp/identd2.log 2>&1 &
pid=$!
sleep 1
reply=$(printf '1,2\r\n' | nc -w 3 127.0.0.1 8000)
echo "$reply" | grep -q ': USERID : UNIX : ident'
kill "$pid"
