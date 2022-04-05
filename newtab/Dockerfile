# Generated from https://github.com/greboid/dockerfiles/blob/master/newtab/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.34.1-r5","apk:ca-certificates":"20211220-r0","apk:expat":"2.4.7-r0","apk:git":"2.34.1-r0","apk:libcrypto1.1":"1.1.1n-r0","apk:libcurl":"7.80.0-r0","apk:libssl1.1":"1.1.1n-r0","apk:musl":"1.2.2-r7","apk:nghttp2-libs":"1.46.0-r0","apk:pcre2":"10.39-r0","apk:zlib":"1.2.12-r0","github:greboid/newtab":"v1.0.0","image:base":"ee8879a758bbe46a6b0b5b9d51a3cd3cb5660106d185b0f1fd981e4475663d08","image:golang":"fbfb7e2408eb1af1b56002c6190df56ffe82b274373db18917c6e4563fd40ca3"}

FROM reg.g5d.dev/golang@sha256:fbfb7e2408eb1af1b56002c6190df56ffe82b274373db18917c6e4563fd40ca3 AS build

ARG TAG="v1.0.0"

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
    git -c advice.detachedHead=false clone --depth=1 -b $TAG --single-branch https://github.com/greboid/newtab /go/src/github.com/greboid/newtab; \
    cd /go/src/github.com/greboid/newtab; \
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    go-licenses save ./... --save_path=/notices;

FROM reg.g5d.dev/base@sha256:ee8879a758bbe46a6b0b5b9d51a3cd3cb5660106d185b0f1fd981e4475663d08

COPY --from=build /go/src/github.com/greboid/newtab/main /newtab
COPY --from=build /notices /notices
EXPOSE 8080
ENTRYPOINT ["/newtab"]
