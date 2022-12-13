# Generated from https://github.com/greboid/dockerfiles/blob/master/goplum/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r2","apk:git":"2.38.2-r0","apk:libcrypto3":"3.0.7-r0","apk:libcurl":"7.86.0-r1","apk:libexpat":"2.5.0-r0","apk:libssl3":"3.0.7-r0","apk:musl":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:pcre2":"10.42-r0","apk:zlib":"1.2.13-r0","github:csmith/goplum":"v0.6.0","image:base":"9af89525fe357fd63d99e49bba4aa94e6e88a7d436500c3588764112f8862109","image:golang":"a28659b5bb93810642775030406b20b18aa645416d66b8c0f265f02cffac0925"}

FROM reg.g5d.dev/golang@sha256:a28659b5bb93810642775030406b20b18aa645416d66b8c0f265f02cffac0925 AS build

ARG TAG="v0.6.0"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r9\
        busybox=1.35.0-r29\
        busybox-binsh=1.35.0-r29\
        ca-certificates=20220614-r2\
        git=2.38.2-r0\
        libcrypto3=3.0.7-r0\
        libcurl=7.86.0-r1\
        libexpat=2.5.0-r0\
        libssl3=3.0.7-r0\
        musl=1.2.3-r4\
        nghttp2-libs=1.51.0-r0\
        pcre2=10.42-r0\
        zlib=1.2.13-r0\
        ; \
    git -c advice.detachedHead=false clone --depth=1 -b $TAG --single-branch https://github.com/csmith/goplum /go/src/github.com/csmith/goplum; \
    cd /go/src/github.com/csmith/goplum; \
    #Build the plugins (all, its easier)
    for plugin in $(ls plugins); do go build -o $plugin.so -buildmode=plugin ./plugins/$plugin/cmd; done; \
    #Build goplum
    go build -o main ./cmd/goplum; \
    go-licenses save ./... --save_path=/notices;

FROM reg.g5d.dev/base@sha256:9af89525fe357fd63d99e49bba4aa94e6e88a7d436500c3588764112f8862109

COPY --from=build /go/src/github.com/csmith/goplum/main /goplum
COPY --from=build /go/src/github.com/csmith/goplum/*.so /plugins/
COPY --from=build /notices /notices
EXPOSE 8080
ENTRYPOINT ["/goplum"]
