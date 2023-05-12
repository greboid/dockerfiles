# Generated from https://github.com/greboid/dockerfiles/blob/master/rust/Dockerfile.gotpl
# BOM: {"github:rust-lang/rust":"1.69.0","github:rust-lang/rustup":"1.26.0","image:alpine":"5b99fedf90b1c3a4d68ed2cac55f46916b91fab77e98f59de05998368fcbf203"}

FROM reg.g5d.dev/alpine@sha256:5b99fedf90b1c3a4d68ed2cac55f46916b91fab77e98f59de05998368fcbf203 AS rustbuild

ARG RUSTTAG="1.69.0"
ARG RUSTUPTAG="1.26.0"

ENV RUSTUP_HOME=/usr/local/rustup \
    CARGO_HOME=/usr/local/cargo \
    PATH=/usr/local/cargo/bin;$PATH \
    RUST_VERSION=${RUSTTAG}

RUN set -eux; \
apk add curl; \
curl -sfLo rustup-init https://static.rust-lang.org/rustup/archive/${RUSTUPTAG}/x86_64-unknown-linux-musl/rustup-init; \
curl -sfLo rustup-init.sha256 https://static.rust-lang.org/rustup/archive/${RUSTUPTAG}/x86_64-unknown-linux-musl/rustup-init.sha256

RUN set -eux; \
apk add curl gpg gpg-agent musl-utils bash libgcc; \
echo $(cat rustup-init.sha256 | awk '{print $1}') *rustup-init | sha256sum -wc -; \
chmod +x rustup-init; \
https://static.rust-lang.org/rustup/archive/1.26.0/x86_64-unknown-linux-musl/rustup-init; \
rm rustup-init*
chmod -R a+w $RUSTUP_HOME $CARGO_HOME
