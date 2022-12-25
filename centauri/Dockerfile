# Generated from https://github.com/greboid/dockerfiles/blob/master/centauri/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r3","apk:git":"2.38.2-r0","apk:libcrypto3":"3.0.7-r0","apk:libcurl":"7.87.0-r0","apk:libexpat":"2.5.0-r0","apk:libssl3":"3.0.7-r0","apk:musl":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:pcre2":"10.42-r0","apk:zlib":"1.2.13-r0","image:base":"ffd04efedafc54f5392c35a1045604c12a4329e5780407d5fe08dfe495d411b1","image:golang":"330555a83d0c2b3834d902e9186b7c2f56613d2c62bad1d6a884c103dd05b698"}

FROM reg.g5d.dev/golang@sha256:330555a83d0c2b3834d902e9186b7c2f56613d2c62bad1d6a884c103dd05b698 AS build

ARG REPO="csmith/centauri"
ARG TAG=4c2e02c5415926b85a5cf5bc75c4a23847b69525

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r9\
        busybox=1.35.0-r29\
        busybox-binsh=1.35.0-r29\
        ca-certificates=20220614-r3\
        git=2.38.2-r0\
        libcrypto3=3.0.7-r0\
        libcurl=7.87.0-r0\
        libexpat=2.5.0-r0\
        libssl3=3.0.7-r0\
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

FROM reg.g5d.dev/base@sha256:ffd04efedafc54f5392c35a1045604c12a4329e5780407d5fe08dfe495d411b1

COPY --from=build --chown=65532:65532 /rootfs/ /
EXPOSE 8080
EXPOSE 8443
ENTRYPOINT ["/centauri"]
