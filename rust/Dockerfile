# Generated from https://github.com/greboid/dockerfiles/blob/master/rust/Dockerfile.gotpl
# BOM: {"github:rust-lang/rust":"1.66.0","image:alpine":"b49a27cfed4e1552dec9027ce598b993cf7a036b1cf7de569aeb62fddac5d70c"}

FROM reg.g5d.dev/alpine@sha256:b49a27cfed4e1552dec9027ce598b993cf7a036b1cf7de569aeb62fddac5d70c AS rustbuild

ARG RUSTTAG="1.66.0"

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
