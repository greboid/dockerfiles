#!/bin/sh
# -trimpath removes the compiled-in GOROOT. v1.6.0 otherwise treats
# every absolute source path as stdlib and silently emits no notices.
# Resolve GOROOT from the installed Go toolchain before starting it.
mkdir -p "${DESTDIR}/usr/libexec"
mv "${DESTDIR}/usr/bin/go-licenses" "${DESTDIR}/usr/libexec/go-licenses"
install -Dm 0755 go-licenses-wrapper.sh "${DESTDIR}/usr/bin/go-licenses"
