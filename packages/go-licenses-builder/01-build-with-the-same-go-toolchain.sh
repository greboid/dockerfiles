#!/bin/sh
# Melange's runner sets its own PATH; select the legacy toolchain here.
export PATH="/go/bin:/usr/local/go/bin:$PATH"
mkdir -p "${DESTDIR}/go/bin"
go build -trimpath -ldflags='-s -w -buildid=' -o "${DESTDIR}/go/bin/go-licenses.bin" ./
install -Dm 0755 go-licenses-wrapper.sh "${DESTDIR}/go/bin/go-licenses"
# Match the shared GOPATH directory permissions owned by go.
chmod 0777 "${DESTDIR}/go" "${DESTDIR}/go/bin"
