#!/bin/sh
# TOKEN and URL are required; without them it prints usage and exits 0.
out=$(/usr/bin/webhooked 2>&1)
code=$?
test "$code" = 0
echo "$out" | grep -q "Usage of /usr/bin/webhooked:"
