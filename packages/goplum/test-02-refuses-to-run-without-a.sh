#!/bin/sh
set +e
out=$(/usr/bin/goplum 2>&1)
code=$?
set -e
test "$code" = 1
echo "$out" | grep -q "Unable to read config"
