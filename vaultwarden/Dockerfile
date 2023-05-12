# Generated from https://github.com/greboid/dockerfiles/blob/master/vaultwarden/Dockerfile.gotpl
# BOM: {"github:dani-garcia/vaultwarden":"1.28.1","image:base":"a98010a83248020eabd68f86ab1c1711d43693c5bb2e0db86e227b37af2e7ec1","image:rust":"3a8243c4f25279c44ea428cbd7b07d2870ca570cc4527a1302697d95c437cd9d"}

FROM reg.g5d.dev/rust@sha256:3a8243c4f25279c44ea428cbd7b07d2870ca570cc4527a1302697d95c437cd9d as server

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
