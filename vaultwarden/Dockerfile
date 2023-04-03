# Generated from https://github.com/greboid/dockerfiles/blob/master/vaultwarden/Dockerfile.gotpl
# BOM: {"github:dani-garcia/vaultwarden":"1.28.1","image:base":"ef8b65fda5d776551a87f970b02ebd1aa67f98c60b68637d910429a4e18be9b9","image:rust":"0b9baf158eb5a237110c44bf770949bb10247b64510ef30899dd18c8c535c250"}

FROM reg.g5d.dev/rust@sha256:0b9baf158eb5a237110c44bf770949bb10247b64510ef30899dd18c8c535c250 as server

ARG VAULTTAG="1.28.1"

RUN set -eux; \
apk add git; \
git clone --depth=1 -b $VAULTTAG --single-branch https://github.com/dani-garcia/vaultwarden /src; \
cd /src

WORKDIR /src

RUN set -eux; \
apk add gcc musl-dev openssl-dev; \
cargo build --features sqlite,enable_mimalloc --release

FROM reg.g5d.dev/base@sha256:ef8b65fda5d776551a87f970b02ebd1aa67f98c60b68637d910429a4e18be9b9

COPY --from=server /src/target/release/vaultwarden /

ENTRYPOINT ["/vaultwarden"]
