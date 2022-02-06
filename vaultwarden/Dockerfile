# Generated from https://github.com/greboid/dockerfiles/blob/master/vaultwarden/Dockerfile.gotpl
# BOM: {"github:dani-garcia/vaultwarden":"1.24.0","github:rust-lang/rust":"1.58.1","image:alpine":"ef95559f475cdf56e55db71c168f1616d497e2215bbd9f6bc63398036ae546a1"}

FROM reg.g5d.dev/alpine@sha256:ef95559f475cdf56e55db71c168f1616d497e2215bbd9f6bc63398036ae546a1 AS rustbuild

ARG RUSTTAG="1.58.1"
ARG VAULTTAG="1.24.0"

RUN set -eux; \
apk add curl; \
curl -sfLo rust.tar.gz https://static.rust-lang.org/dist/rust-${RUSTTAG}-x86_64-unknown-linux-musl.tar.gz; \
curl -sfLo rust.tar.gz.asc https://static.rust-lang.org/dist/rust-${RUSTTAG}-x86_64-unknown-linux-musl.tar.gz.asc

RUN set -eux; \
apk add curl gpg gpg-agent musl-utils bash libgcc; \
curl https://keybase.io/rust/pgp_keys.asc | gpg --import; \
gpg --batch --verify rust.tar.gz.asc rust.tar.gz; \
tar -C /tmp -xf rust.tar.gz; \
mv /tmp/rust-* /tmp/rust; \
cd /tmp/rust; \
/tmp/rust/install.sh --verbose

ENV VAULTTAG=async

RUN set -eux; \
apk add git; \
git clone --depth=1 -b $VAULTTAG --single-branch https://github.com/dani-garcia/vaultwarden /src; \
cd /src

WORKDIR /src

ENV RUSTFLAGS=$RUSTFLAGS" -L native=/usr/lib -l static=clang -l static=icui18n -l static=icuuc -l static=icudata -l static=stdc++"
ENV RUSTFLAGS=$RUSTFLAGS" -C target-feature=+crt-static -C link-self-contained=yes"
ENV RUSTC_BOOTSTRAP=1

RUN set -eux; \
apk add gcc musl-dev openssl-dev; \
cargo build --features sqlite --release
