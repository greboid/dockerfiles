# Generated from https://github.com/greboid/dockerfiles/blob/master/legoergo/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r4","apk:git":"2.38.4-r1","apk:libcrypto3":"3.0.8-r3","apk:libcurl":"7.88.1-r1","apk:libexpat":"2.5.0-r0","apk:libssl3":"3.0.8-r3","apk:musl":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:pcre2":"10.42-r0","apk:zlib":"1.2.13-r0","github:csmith/certwrapper":"v4.1.0","github:ergochat/ergo":"v2.11.1","image:base":"ef8b65fda5d776551a87f970b02ebd1aa67f98c60b68637d910429a4e18be9b9","image:golang":"20b7658093fc4e7748edddd67a45e19fe467d846b9fcf3ceaa37a5f84f672f14"}

FROM reg.g5d.dev/golang@sha256:20b7658093fc4e7748edddd67a45e19fe467d846b9fcf3ceaa37a5f84f672f14 AS build

ARG CERTTAG="v4.1.0"
ARG ERGOTAG="v2.11.1"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r9 \
        busybox=1.35.0-r29 \
        busybox-binsh=1.35.0-r29 \
        ca-certificates=20220614-r4 \
        git=2.38.4-r1 \
        libcrypto3=3.0.8-r3 \
        libcurl=7.88.1-r1 \
        libexpat=2.5.0-r0 \
        libssl3=3.0.8-r3 \
        musl=1.2.3-r4 \
        nghttp2-libs=1.51.0-r0 \
        pcre2=10.42-r0 \
        zlib=1.2.13-r0 \
        ; \
    git clone --depth=1 -b $CERTTAG --single-branch https://github.com/csmith/certwrapper /go/src/github.com/csmith/certwrapper; \
    cd /go/src/github.com/csmith/certwrapper; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -tags lego_httpreq -o main .; \
    git clone --depth=1 -b $ERGOTAG --single-branch https://github.com/ergochat/ergo /go/src/github.com/ergochat/ergo; \
    cd /go/src/github.com/ergochat/ergo; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    go-licenses save ./... --save_path=/notices;

FROM reg.g5d.dev/base@sha256:ef8b65fda5d776551a87f970b02ebd1aa67f98c60b68637d910429a4e18be9b9

COPY --from=build /go/src/github.com/csmith/certwrapper/main /certwrapper
COPY --from=build /go/src/github.com/ergochat/ergo/main /ergo
COPY --from=build /go/src/github.com/ergochat/ergo/languages /ircd-bin/languages
COPY --from=build /notices /notices
CMD ["/certwrapper", "/ergo", "run", "--conf", "/ircd/ircd.yaml"]
