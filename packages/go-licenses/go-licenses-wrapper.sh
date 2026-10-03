#!/bin/sh
set -eu
GOROOT="${GOROOT:-$(go env GOROOT)}"
export GOROOT
exec /usr/libexec/go-licenses "$@"
