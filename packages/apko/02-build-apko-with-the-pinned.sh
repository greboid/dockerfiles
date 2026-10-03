#!/bin/sh
# The melange runner resets PATH: put the custom toolchain on it explicitly
# instead of relying on environment.environment.PATH.
export PATH="/usr/local/go/bin:$PATH"
export GOROOT=/usr/local/go
set -eu
go build -trimpath -o apko .
install -Dm755 apko "${DESTDIR}/usr/bin/apko"
