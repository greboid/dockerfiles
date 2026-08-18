#!/bin/sh
set -e

IMAGE="${1:-reg.g5d.dev/git}"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

podman run --rm --entrypoint git "$IMAGE" --version | grep -q "git version"

# The remote-https helper must resolve inside the compiled-in exec path
test "$(podman run --rm --entrypoint git "$IMAGE" --exec-path)" = "/usr/libexec/git-core"

# Cloning over HTTPS must work (CA bundle is mounted in as the image is from scratch)
podman run --rm --entrypoint git \
    -v /etc/ssl/certs/ca-certificates.crt:/etc/ssl/certs/ca-certificates.crt:ro \
    -e SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt \
    -v "$TMP":/target \
    "$IMAGE" clone --depth=1 https://github.com/octocat/Hello-World.git /target/hello

test -f "$TMP/hello/README"
