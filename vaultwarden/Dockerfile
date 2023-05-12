# Generated from https://github.com/greboid/dockerfiles/blob/master/vaultwarden/Dockerfile.gotpl
# BOM: {"github:dani-garcia/vaultwarden":"1.28.1","image:base":"a98010a83248020eabd68f86ab1c1711d43693c5bb2e0db86e227b37af2e7ec1","image:rust":"26350c82c19d33be0908d83c50dfce927a4249fba7cc252cf6929ee0d61a1c23"}

FROM reg.g5d.dev/rust@sha256:26350c82c19d33be0908d83c50dfce927a4249fba7cc252cf6929ee0d61a1c23 as server

ARG VAULTTAG="1.28.1"

RUN set -eux; \
apk add git; \
git clone --depth=1 -b $VAULTTAG --single-branch https://github.com/dani-garcia/vaultwarden /src; \
cd /src

WORKDIR /src

RUN set -eux; \
apk add gcc musl-dev openssl-dev; \
cargo build --features sqlite,enable_mimalloc --release

FROM reg.g5d.dev/base@sha256:a98010a83248020eabd68f86ab1c1711d43693c5bb2e0db86e227b37af2e7ec1

COPY --from=server /src/target/release/vaultwarden /

ENTRYPOINT ["/vaultwarden"]
