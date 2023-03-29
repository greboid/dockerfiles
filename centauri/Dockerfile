# Generated from https://github.com/greboid/dockerfiles/blob/master/centauri/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r4","apk:git":"2.38.4-r1","apk:libcrypto3":"3.0.8-r2","apk:libcurl":"7.88.1-r1","apk:libexpat":"2.5.0-r0","apk:libssl3":"3.0.8-r2","apk:musl":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:pcre2":"10.42-r0","apk:zlib":"1.2.13-r0","image:base":"ef8b65fda5d776551a87f970b02ebd1aa67f98c60b68637d910429a4e18be9b9","image:golang":"f37fa1d54487659f3ef72dd9084dacfef3a8eae9a32410843f983da55f7eb16b"}

FROM reg.g5d.dev/golang@sha256:f37fa1d54487659f3ef72dd9084dacfef3a8eae9a32410843f983da55f7eb16b AS build

ARG REPO="csmith/centauri"
ARG TAG=4c2e02c5415926b85a5cf5bc75c4a23847b69525

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r9\
        busybox=1.35.0-r29\
        busybox-binsh=1.35.0-r29\
        ca-certificates=20220614-r4\
        git=2.38.4-r1\
        libcrypto3=3.0.8-r2\
        libcurl=7.88.1-r1\
        libexpat=2.5.0-r0\
        libssl3=3.0.8-r2\
        musl=1.2.3-r4\
        nghttp2-libs=1.51.0-r0\
        pcre2=10.42-r0\
        zlib=1.2.13-r0\
        ; \
    git clone -b master --single-branch https://github.com/$REPO /go/src; \
    cd /go/src; \
    git checkout $TAG; \
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -tags lego_httpreq -o main ./cmd/centauri; \
    chmod +x /go/src/main; \
    mkdir -p /rootfs/home/nonroot/.config; \
    mkdir -p /rootfs/data; \
    mv /go/src/main /rootfs/centauri

FROM reg.g5d.dev/base@sha256:ef8b65fda5d776551a87f970b02ebd1aa67f98c60b68637d910429a4e18be9b9

COPY --from=build --chown=65532:65532 /rootfs/ /
EXPOSE 8080
EXPOSE 8443
ENTRYPOINT ["/centauri"]
