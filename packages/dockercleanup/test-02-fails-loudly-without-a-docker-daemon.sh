#!/bin/sh
# With DURATION below one minute it runs a single pass, so a missing
# daemon surfaces immediately instead of hanging.
set +e
out=$(DOCKER_HOST=tcp://127.0.0.1:9 /usr/bin/dockercleanup 2>&1)
code=$?
set -e
test "$code" = 1
echo "$out" | grep -q "unable to clean perform cleanup"
