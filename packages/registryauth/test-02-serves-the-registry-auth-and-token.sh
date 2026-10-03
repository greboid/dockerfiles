#!/bin/sh
# The login password is "smoke-pass"; the hash pins the test.
USERS='{smoke: $2b$12$QPcRNX6Epm5/L5IMLs8/de.M9paX.qVTLEZ9e9StqkH9/NjeccAEm}' \
  DATA_DIR=/tmp/ra-data \
  /usr/bin/registryauth >/tmp/registryauth.log 2>&1 &
pid=$!
# Self-signed certificate generation takes a moment.
for _ in $(seq 1 20); do
  code=$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:8080/v2/ || true)
  [ "$code" = 401 ] && break
  sleep 1
done
test "$code" = 401
# Valid credentials yield a token; invalid ones are rejected when no
# scope is requested (upstream behavior).
token=$(curl -sf -u smoke:smoke-pass \
  'http://127.0.0.1:8080/auth?service=Registry')
echo "$token" | grep -q access_token
! curl -sf -u smoke:wrong-password \
  'http://127.0.0.1:8080/auth?service=Registry' >/dev/null
kill "$pid"
grep -q "Starting registry" /tmp/registryauth.log
