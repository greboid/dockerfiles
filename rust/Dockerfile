# Generated from https://github.com/greboid/dockerfiles/blob/master/rust/Dockerfile.gotpl
# BOM: {"github:rust-lang/rust":"1.68.0","image:alpine":"e3385453363598236b7e826530c55b70d4a1a3e53931d4e7b675d47c4c5c8f8d"}

FROM reg.g5d.dev/alpine@sha256:e3385453363598236b7e826530c55b70d4a1a3e53931d4e7b675d47c4c5c8f8d AS rustbuild

ARG RUSTTAG="1.68.0"

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
