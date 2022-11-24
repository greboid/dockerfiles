# Generated from https://github.com/greboid/dockerfiles/blob/master/puzzles/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r2","apk:git":"2.38.1-r0","apk:libcrypto3":"3.0.7-r0","apk:libcurl":"7.86.0-r1","apk:libexpat":"2.5.0-r0","apk:libssl3":"3.0.7-r0","apk:musl":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:pcre2":"10.40-r0","apk:zlib":"1.2.13-r0","github:greboid/puzzles":"v1.0.1","image:base":"9af89525fe357fd63d99e49bba4aa94e6e88a7d436500c3588764112f8862109","image:golang":"3277ef0b2cb0544aae554b8b9377d7af0a24c80bef792f613bbd6d176498c23b"}

FROM reg.g5d.dev/golang@sha256:3277ef0b2cb0544aae554b8b9377d7af0a24c80bef792f613bbd6d176498c23b AS build

ARG TAG="v1.0.1"

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
    git -c advice.detachedHead=false clone --depth=1 -b $TAG --single-branch https://github.com/greboid/puzzles /go/src/github.com/greboid/puzzles; \
    cd /go/src/github.com/greboid/puzzles; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    go-licenses save ./... --save_path=/notices;

FROM reg.g5d.dev/base@sha256:9af89525fe357fd63d99e49bba4aa94e6e88a7d436500c3588764112f8862109

COPY --from=build /go/src/github.com/greboid/puzzles/main /puzzles
COPY --from=build /go/src/github.com/greboid/puzzles/wordlists/. /app/wordlists
COPY --from=build /notices /notices
EXPOSE 8080
ENTRYPOINT ["/puzzles"]
