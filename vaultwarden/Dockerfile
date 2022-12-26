# Generated from https://github.com/greboid/dockerfiles/blob/master/vaultwarden/Dockerfile.gotpl
# BOM: {"github:dani-garcia/vaultwarden":"1.27.0","image:base":"ffd04efedafc54f5392c35a1045604c12a4329e5780407d5fe08dfe495d411b1","image:rust":"813ea1254f030b56b7b180001985391b3dd49a036a8bffdf9884ba39ff8a67f5"}

FROM reg.g5d.dev/rust@sha256:813ea1254f030b56b7b180001985391b3dd49a036a8bffdf9884ba39ff8a67f5 as server

ARG VAULTTAG="1.27.0"

RUN set -eux; \
apk add git; \
git clone --depth=1 -b $VAULTTAG --single-branch https://github.com/dani-garcia/vaultwarden /src; \
cd /src

WORKDIR /src

RUN set -eux; \
apk add gcc musl-dev openssl-dev; \
cargo build --features sqlite,enable_mimalloc --release

FROM reg.g5d.dev/base@sha256:ffd04efedafc54f5392c35a1045604c12a4329e5780407d5fe08dfe495d411b1

COPY --from=server /src/target/release/vaultwarden /

ENTRYPOINT ["/vaultwarden"]
