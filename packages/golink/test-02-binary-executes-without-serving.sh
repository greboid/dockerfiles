#!/bin/sh
# Usage output proves the static binary runs end to end. No listener
# is started: the packaged service behaviour is exercised by the
# image tests, not by packaging.
/usr/bin/golink -h 2>&1 | grep -q dev-listen
