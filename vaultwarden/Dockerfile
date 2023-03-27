# Generated from https://github.com/greboid/dockerfiles/blob/master/vaultwarden/Dockerfile.gotpl
# BOM: {"github:dani-garcia/vaultwarden":"1.28.0","image:base":"6864cad4f0ad8d3dcdf9724a9fd7df80331f9558a60be2c4639ae8243e4fa116","image:rust":"2f4f15dcb35bf80cb75bf574851f1ad595e0188fbbb3596f6179ef136cd086a3"}

FROM reg.g5d.dev/rust@sha256:2f4f15dcb35bf80cb75bf574851f1ad595e0188fbbb3596f6179ef136cd086a3 as server

ARG VAULTTAG="1.28.0"

RUN set -eux; \
apk add git; \
git clone --depth=1 -b $VAULTTAG --single-branch https://github.com/dani-garcia/vaultwarden /src; \
cd /src

WORKDIR /src

RUN set -eux; \
apk add gcc musl-dev openssl-dev; \
cargo build --features sqlite,enable_mimalloc --release

FROM reg.g5d.dev/base@sha256:6864cad4f0ad8d3dcdf9724a9fd7df80331f9558a60be2c4639ae8243e4fa116

COPY --from=server /src/target/release/vaultwarden /

ENTRYPOINT ["/vaultwarden"]
