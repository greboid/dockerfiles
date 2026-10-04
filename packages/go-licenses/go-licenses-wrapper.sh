#!/bin/sh
# Resolve the consumer's selected toolchain, not the collector's baked-in
# GOROOT: otherwise downloaded toolchain stdlib is treated as third-party.
exec env GOROOT="$(go env GOROOT)" /usr/libexec/go-licenses "$@"
