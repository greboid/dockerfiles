#!/bin/sh
# The melange runner resets PATH: put the custom toolchain on it
# explicitly instead of relying on environment.environment.PATH.
export PATH="/usr/local/go/bin:$PATH"
export GOROOT=/usr/local/go
make bin/buildah
install -Dm755 bin/buildah "${DESTDIR}/usr/bin/buildah"
