# Generated from https://github.com/greboid/dockerfiles/blob/master/centauri/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r14","apk:busybox":"1.36.0-r9","apk:busybox-binsh":"1.36.0-r9","apk:ca-certificates":"20230506-r0","apk:git":"2.40.1-r0","apk:libcrypto3":"3.1.0-r4","apk:libcurl":"8.1.0-r0","apk:libexpat":"2.5.0-r1","apk:libidn2":"2.3.4-r1","apk:libssl3":"3.1.0-r4","apk:libunistring":"1.1-r1","apk:musl":"1.2.4-r0","apk:nghttp2-libs":"1.53.0-r0","apk:pcre2":"10.42-r1","apk:zlib":"1.2.13-r1","image:base":"4ddff4468810fe9d9dc3285917d2923c41ae9fb6d1fa2b91c3f1d460cb94427a","image:golang":"866a26e8691570c5e39e43030b288dde9b22387f36f3b58d659af14fcbbf9477"}

FROM reg.g5d.dev/golang@sha256:866a26e8691570c5e39e43030b288dde9b22387f36f3b58d659af14fcbbf9477 AS build

ARG REPO="csmith/centauri"
ARG TAG=4c2e02c5415926b85a5cf5bc75c4a23847b69525

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r14\
        busybox=1.36.0-r9\
        busybox-binsh=1.36.0-r9\
        ca-certificates=20230506-r0\
        git=2.40.1-r0\
        libcrypto3=3.1.0-r4\
        libcurl=8.1.0-r0\
        libexpat=2.5.0-r1\
        libidn2=2.3.4-r1\
        libssl3=3.1.0-r4\
        libunistring=1.1-r1\
        musl=1.2.4-r0\
        nghttp2-libs=1.53.0-r0\
        pcre2=10.42-r1\
        zlib=1.2.13-r1\
        ; \
    git clone -b master --single-branch https://github.com/$REPO /go/src; \
    cd /go/src; \
    git checkout $TAG; \
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -tags lego_httpreq -o main ./cmd/centauri; \
    chmod +x /go/src/main; \
    mkdir -p /rootfs/home/nonroot/.config; \
    mkdir -p /rootfs/data; \
    mv /go/src/main /rootfs/centauri

FROM reg.g5d.dev/base@sha256:4ddff4468810fe9d9dc3285917d2923c41ae9fb6d1fa2b91c3f1d460cb94427a

COPY --from=build --chown=65532:65532 /rootfs/ /
EXPOSE 8080
EXPOSE 8443
ENTRYPOINT ["/centauri"]
