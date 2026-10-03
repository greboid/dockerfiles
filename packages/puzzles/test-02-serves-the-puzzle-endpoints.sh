#!/bin/sh
/usr/bin/puzzles >/tmp/puzzles.log 2>&1 &
pid=$!
for _ in $(seq 1 10); do
  code=$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:8080/ || true)
  [ -n "$code" ] && [ "$code" != 000 ] && break
  sleep 1
done
# / redirects permanently to the public site.
test "$code" = 308
curl -sD /tmp/headers -o /dev/null http://127.0.0.1:8080/
grep -qi '^location: https://puzzles.mdbot.uk' /tmp/headers
# The word endpoints consult the packaged wordlists.
anagram=$(curl -sf 'http://127.0.0.1:8080/anagram?input=teerts')
echo "$anagram" | grep -q '<li>'
# Embedded static assets are served from the binary's root.
curl -sf -o /dev/null http://127.0.0.1:8080/main.js
curl -sf -o /dev/null http://127.0.0.1:8080/flags/GB-flag.webp
test "$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:8080/nope)" = 404
kill "$pid"
