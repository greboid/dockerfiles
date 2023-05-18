# Generated from https://github.com/greboid/dockerfiles/blob/master/vaultwarden/Dockerfile.gotpl
# BOM: {"github:dani-garcia/vaultwarden":"1.28.1","image:base":"4ddff4468810fe9d9dc3285917d2923c41ae9fb6d1fa2b91c3f1d460cb94427a","image:rust":"26350c82c19d33be0908d83c50dfce927a4249fba7cc252cf6929ee0d61a1c23"}

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

FROM reg.g5d.dev/base@sha256:4ddff4468810fe9d9dc3285917d2923c41ae9fb6d1fa2b91c3f1d460cb94427a

COPY --from=server /src/target/release/vaultwarden /

ENTRYPOINT ["/vaultwarden"]
