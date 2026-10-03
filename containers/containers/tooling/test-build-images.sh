#!/bin/sh
# Test loaded golang, rust and alpine-runner images (apko's :local-amd64 tags).
# No publishing or network access. --privileged-runner additionally exercises
# nested Buildah in a disposable privileged container (no host bind mounts).
set -eu

case "${1:-}" in
    '') privileged=false ;;
    --privileged-runner) privileged=true ;;
    *) echo "usage: $0 [--privileged-runner]" >&2; exit 2 ;;
esac

work=$(mktemp -d)
go_image="builder-smoke-go-$$:local"
rust_image="builder-smoke-rust-$$:local"
cleanup() {
    rm -rf "$work"
    docker image rm "$go_image" "$rust_image" >/dev/null 2>&1 || true
}
trap cleanup EXIT

# Explicit commands must not be swallowed by an image entrypoint. These images
# are also expected to support root-owned package installation in child stages.
for image in golang rust alpine-runner; do
    docker run --rm --network none "$image:local-amd64" /bin/sh -ec '
        test "$(id -u)" = 0
        test "$LANG" = en_US.UTF-8
        command -v apk
        apk --version
        test -s /etc/apk/repositories
        test -n "$(find /etc/apk/keys -name "*.pub" -print)"
        git --version
    '
done

mkdir "$work/go" "$work/rust"
cat > "$work/go/Dockerfile" <<'EOF'
FROM golang:local-amd64 AS build
WORKDIR /smoke
COPY . .
RUN test "$(go env GOVERSION)" = go1.27.1 && \
    test "$GOPATH" = /go && test "$GOFLAGS" = -buildvcs=false && \
    test "$GOAMD64" = v1 && test "$GOTELEMETRY" = off && \
    command -v go-licenses && \
    CGO_ENABLED=0 go build -o /smoke/pure pure.go && \
    CGO_ENABLED=1 go build -o /smoke/cgo cgo.go && /smoke/cgo
FROM scratch
COPY --from=build /smoke/pure /smoke
ENTRYPOINT ["/smoke"]
EOF
cat > "$work/go/pure.go" <<'EOF'
package main
import "fmt"
func main() { fmt.Println("go builder ok") }
EOF
cat > "$work/go/cgo.go" <<'EOF'
package main
// #include <stdlib.h>
import "C"
func main() { if C.abs(-42) != 42 { panic("cgo failed") } }
EOF

docker build --network none --pull=false -t "$go_image" "$work/go"
test "$(docker run --rm --network none "$go_image")" = 'go builder ok'

cat > "$work/rust/Dockerfile" <<'EOF'
FROM rust:local-amd64 AS build
WORKDIR /smoke
COPY . .
RUN rustc --version | grep '^rustc 1.98.1 ' && \
    rustup --version | grep '^rustup 1.29.1 ' && \
    test "$CARGO_HOME" = /usr/local/cargo && \
    test "$RUSTUP_HOME" = /usr/local/rustup && \
    cargo build --offline --release && \
    ./target/release/builder-smoke
FROM scratch
COPY --from=build /smoke/target/release/builder-smoke /smoke
ENTRYPOINT ["/smoke"]
EOF
cat > "$work/rust/Cargo.toml" <<'EOF'
[package]
name = "builder-smoke"
version = "0.1.0"
edition = "2021"
EOF
mkdir "$work/rust/src"
printf '%s\n' 'fn main() { println!("rust builder ok"); }' > "$work/rust/src/main.rs"
docker build --network none --pull=false -t "$rust_image" "$work/rust"
test "$(docker run --rm --network none "$rust_image")" = 'rust builder ok'

docker run --rm --network none alpine-runner:local-amd64 /bin/sh -ec '
    buildah --version | grep "buildah version 1.45.1"
    for tool in fuse-overlayfs curl wget openssl gpg; do
        command -v "$tool"
    done
    apk info -e netavark
    config=${CONTAINERS_STORAGE_CONF:-/etc/containers/storage.conf}
    grep "driver = \"overlay\"" "$config"
    grep "/usr/bin/fuse-overlayfs" "$config"
'

if "$privileged"; then
    docker run --rm --privileged --network none alpine-runner:local-amd64 /bin/sh -ec '
        mkdir /tmp/context
        printf "nested build ok\n" > /tmp/context/probe
        printf "FROM scratch\nCOPY probe /probe\n" > /tmp/context/Containerfile
        buildah bud --network host --isolation chroot -t localhost/builder-smoke /tmp/context
        ctr=$(buildah from localhost/builder-smoke)
        root=$(buildah mount "$ctr")
        test "$(cat "$root/probe")" = "nested build ok"
        buildah unmount "$ctr"
        buildah rm "$ctr"
    '
else
    echo 'Nested Buildah check skipped; opt in with --privileged-runner.'
fi

echo 'Build-image smoke checks passed.'
