#!/bin/sh
# Missing AUTHTOKEN prints usage and exits cleanly; nothing is
# contacted.
/usr/bin/githubmirror >/tmp/githubmirror.log 2>&1
test $? -eq 0
grep -q "authtoken" /tmp/githubmirror.log
