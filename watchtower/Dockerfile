# Generated from https://github.com/greboid/dockerfiles/blob/master/watchtower/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r2","apk:git":"2.38.1-r0","apk:libcrypto3":"3.0.7-r0","apk:libcurl":"7.86.0-r1","apk:libexpat":"2.5.0-r0","apk:libssl3":"3.0.7-r0","apk:musl":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:pcre2":"10.40-r0","apk:zlib":"1.2.13-r0","github:containrrr/watchtower":"v1.5.1","image:base":"9af89525fe357fd63d99e49bba4aa94e6e88a7d436500c3588764112f8862109","image:golang":"4642bb640987c257a4b3581743fe1cc73f3f9b87252b9f1910be92368656c1e4"}

FROM reg.g5d.dev/golang@sha256:4642bb640987c257a4b3581743fe1cc73f3f9b87252b9f1910be92368656c1e4 AS build

ARG TAG="v1.5.1"

COPY disable-slack.patch /

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r9\
        busybox=1.35.0-r29\
        busybox-binsh=1.35.0-r29\
        ca-certificates=20220614-r2\
        git=2.38.1-r0\
        libcrypto3=3.0.7-r0\
        libcurl=7.86.0-r1\
        libexpat=2.5.0-r0\
        libssl3=3.0.7-r0\
        musl=1.2.3-r4\
        nghttp2-libs=1.51.0-r0\
        pcre2=10.40-r0\
        zlib=1.2.13-r0\
        ; \
    git -c advice.detachedHead=false clone --depth=1 -b $TAG --single-branch https://github.com/containrrr/watchtower /go/src/github.com/containrrr/watchtower; \
    cd /go/src/github.com/containrrr/watchtower; \
    git apply /disable-slack.patch; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -ldflags "-X github.com/containrrr/watchtower/internal/meta.Version=$TAG" -o main .; \
    go-licenses save ./... --save_path=/notices;

FROM reg.g5d.dev/base@sha256:9af89525fe357fd63d99e49bba4aa94e6e88a7d436500c3588764112f8862109

COPY --from=build /go/src/github.com/containrrr/watchtower/main /watchtower
#COPY --from=build /notices /notices
ENTRYPOINT ["/watchtower"]
