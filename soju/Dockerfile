# Generated from https://github.com/greboid/dockerfiles/blob/master/soju/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r2","apk:git":"2.38.2-r0","apk:libcrypto3":"3.0.7-r0","apk:libcurl":"7.86.0-r1","apk:libexpat":"2.5.0-r0","apk:libssl3":"3.0.7-r0","apk:musl":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:patch":"2.7.6-r8","apk:pcre2":"10.42-r0","apk:zlib":"1.2.13-r0","github:csmith/certwrapper":"v4.1.0","github:emersion/soju":"v0.5.2","image:base":"9af89525fe357fd63d99e49bba4aa94e6e88a7d436500c3588764112f8862109","image:golang":"4b4ba7b1f9b635c33fe5c6abc66075d02ee7bd1e67d62421169dbcd76d8d81bf"}

FROM reg.g5d.dev/golang@sha256:4b4ba7b1f9b635c33fe5c6abc66075d02ee7bd1e67d62421169dbcd76d8d81bf AS build

ARG SOJUTAG="v0.5.2"
ARG CERTTAG="v4.1.0"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r9 \
        busybox=1.35.0-r29 \
        busybox-binsh=1.35.0-r29 \
        ca-certificates=20220614-r2 \
        git=2.38.2-r0 \
        libcrypto3=3.0.7-r0 \
        libcurl=7.86.0-r1 \
        libexpat=2.5.0-r0 \
        libssl3=3.0.7-r0 \
        musl=1.2.3-r4 \
        nghttp2-libs=1.51.0-r0 \
        patch=2.7.6-r8 \
        pcre2=10.42-r0 \
        zlib=1.2.13-r0 \
        ; \
    git -c advice.detachedHead=false clone --depth=1 -b $SOJUTAG --single-branch https://github.com/emersion/soju /go/src/github.com/emersion/soju; \
    cd /go/src/github.com/emersion/soju; \
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o soju ./cmd/soju; \
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o sojuctl ./cmd/sojuctl; \
    git -c advice.detachedHead=false clone --depth=1 -b $CERTTAG --single-branch https://github.com/csmith/certwrapper /go/src/github.com/csmith/certwrapper; \
    cd /go/src/github.com/csmith/certwrapper;\
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -tags lego_httpreq -o certwrapper .; \
    go-licenses save ./... --save_path=/notices; \
    mkdir /data

FROM reg.g5d.dev/base@sha256:9af89525fe357fd63d99e49bba4aa94e6e88a7d436500c3588764112f8862109

COPY --from=build /go/src/github.com/emersion/soju/soju /soju
COPY --from=build /go/src/github.com/emersion/soju/sojuctl /sojuctl
COPY --from=build /go/src/github.com/csmith/certwrapper/certwrapper /certwrapper
COPY --from=build /notices /notices
COPY --from=build --chown=65535:65535 /data /data
ENTRYPOINT ["/certwrapper", "/soju"]
