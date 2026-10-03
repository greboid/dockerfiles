#!/bin/sh
/usr/bin/redis /redis.conf >/tmp/redis.log 2>&1 &
pid=$!
sleep 1
reply=$(printf 'PING\r\nQUIT\r\n' | nc 127.0.0.1 6379)
printf '%s\n' "$reply" | grep -q '+PONG'
