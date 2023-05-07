# Generated from https://github.com/greboid/dockerfiles/blob/master/centauri/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20230506-r0","apk:git":"2.38.5-r0","apk:libcrypto3":"3.0.8-r4","apk:libcurl":"8.0.1-r0","apk:libexpat":"2.5.0-r0","apk:libssl3":"3.0.8-r4","apk:musl":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:pcre2":"10.42-r0","apk:zlib":"1.2.13-r0","image:base":"2cf8662a916097fbef5fc9aa384be213a8bf229277931807b092431c9c77e90f","image:golang":"5bb7e1c27710c2d0390e6b1456c52ee9d1702cb5e6c4b54cad653b6d1cf7deb4"}

FROM reg.g5d.dev/golang@sha256:5bb7e1c27710c2d0390e6b1456c52ee9d1702cb5e6c4b54cad653b6d1cf7deb4 AS build

ARG REPO="csmith/centauri"
ARG TAG=4c2e02c5415926b85a5cf5bc75c4a23847b69525

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r9\
        busybox=1.35.0-r29\
        busybox-binsh=1.35.0-r29\
        ca-certificates=20230506-r0\
        git=2.38.5-r0\
        libcrypto3=3.0.8-r4\
        libcurl=8.0.1-r0\
        libexpat=2.5.0-r0\
        libssl3=3.0.8-r4\
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

FROM reg.g5d.dev/base@sha256:2cf8662a916097fbef5fc9aa384be213a8bf229277931807b092431c9c77e90f

COPY --from=build --chown=65532:65532 /rootfs/ /
EXPOSE 8080
EXPOSE 8443
ENTRYPOINT ["/centauri"]
