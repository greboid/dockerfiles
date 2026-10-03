#!/bin/sh
bootstrap="$(go env GOROOT)"
(cd src && GOROOT_BOOTSTRAP="$bootstrap" GOROOT_FINAL=/usr/local/go GOHOSTOS=linux GOHOSTARCH=amd64 ./make.bash)
GOROOT="$PWD" ./bin/go install std
