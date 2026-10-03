#!/bin/sh
# Exercise the signing machinery end to end: generate a key pair.
d=$(mktemp -d)
melange keygen "$d/smoke"
test -s "$d/smoke"
test -s "$d/smoke.pub"
