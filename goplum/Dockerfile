# Generated from https://github.com/greboid/dockerfiles/blob/master/goplum/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.34.1-r5","apk:ca-certificates":"20211220-r0","apk:expat":"2.4.7-r0","apk:git":"2.34.1-r0","apk:libcrypto1.1":"1.1.1n-r0","apk:libcurl":"7.80.0-r0","apk:libssl1.1":"1.1.1n-r0","apk:musl":"1.2.2-r7","apk:nghttp2-libs":"1.46.0-r0","apk:pcre2":"10.39-r0","apk:zlib":"1.2.12-r0","github:csmith/goplum":"v0.6.0","image:base":"ee8879a758bbe46a6b0b5b9d51a3cd3cb5660106d185b0f1fd981e4475663d08","image:golang":"985ad6697bdc3be5fb95c4a7bec677bf221e17b03fef363e04e4cc55d480c8b7"}

FROM reg.g5d.dev/golang@sha256:985ad6697bdc3be5fb95c4a7bec677bf221e17b03fef363e04e4cc55d480c8b7 AS build

ARG TAG="v0.6.0"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r5\
        busybox=1.34.1-r5\
        ca-certificates=20211220-r0\
        expat=2.4.7-r0\
        git=2.34.1-r0\
        libcrypto1.1=1.1.1n-r0\
        libcurl=7.80.0-r0\
        libssl1.1=1.1.1n-r0\
        musl=1.2.2-r7\
        nghttp2-libs=1.46.0-r0\
        pcre2=10.39-r0\
        zlib=1.2.12-r0\
        ; \
    git -c advice.detachedHead=false clone --depth=1 -b $TAG --single-branch https://github.com/csmith/goplum /go/src/github.com/csmith/goplum; \
    cd /go/src/github.com/csmith/goplum; \
    #Build the plugins (all, its easier)
    for plugin in $(ls plugins); do go build -o $plugin.so -buildmode=plugin ./plugins/$plugin/cmd; done; \
    #Build goplum
    go build -o main ./cmd/goplum; \
    go-licenses save ./... --save_path=/notices;

FROM reg.g5d.dev/base@sha256:ee8879a758bbe46a6b0b5b9d51a3cd3cb5660106d185b0f1fd981e4475663d08

COPY --from=build /go/src/github.com/csmith/goplum/main /goplum
COPY --from=build /go/src/github.com/csmith/goplum/*.so /plugins/
COPY --from=build /notices /notices
EXPOSE 8080
ENTRYPOINT ["/goplum"]
