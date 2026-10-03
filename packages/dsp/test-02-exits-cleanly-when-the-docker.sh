#!/bin/sh
# The socket check happens before the proxy starts; upstream treats a
# missing socket as a clean exit.
/usr/bin/dsp >/tmp/dsp.log 2>&1
test $? -eq 0
grep -q "socket does not exist" /tmp/dsp.log
