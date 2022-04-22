# Generated from https://github.com/greboid/dockerfiles/blob/master/centauri/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.34.1-r5","apk:ca-certificates":"20211220-r0","apk:expat":"2.4.7-r0","apk:git":"2.34.2-r0","apk:libcrypto1.1":"1.1.1n-r0","apk:libcurl":"7.80.0-r0","apk:libssl1.1":"1.1.1n-r0","apk:musl":"1.2.2-r7","apk:nghttp2-libs":"1.46.0-r0","apk:pcre2":"10.39-r0","apk:zlib":"1.2.12-r0","github:csmith/centauri":"v0.1.0","image:base":"ee8879a758bbe46a6b0b5b9d51a3cd3cb5660106d185b0f1fd981e4475663d08","image:golang":"9bdb14312a9c1387e49536c145075bc1d52c93324bf8d82eef36ba2782ff1eb2"}

FROM reg.g5d.dev/golang@sha256:9bdb14312a9c1387e49536c145075bc1d52c93324bf8d82eef36ba2782ff1eb2 AS build

ARG REPO="csmith/centauri"
ARG TAG="v0.1.0"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r5\
        busybox=1.34.1-r5\
        ca-certificates=20211220-r0\
        expat=2.4.7-r0\
        git=2.34.2-r0\
        libcrypto1.1=1.1.1n-r0\
        libcurl=7.80.0-r0\
        libssl1.1=1.1.1n-r0\
        musl=1.2.2-r7\
        nghttp2-libs=1.46.0-r0\
        pcre2=10.39-r0\
        zlib=1.2.12-r0\
        ; \
    git -c advice.detachedHead=false clone --depth=1 -b $TAG --single-branch https://github.com/$REPO /go/src; \
    cd /go/src; \
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -tags lego_httpreq -o main ./cmd/centauri/main.go; \
    chmod +x /go/src/main;

FROM reg.g5d.dev/base@sha256:ee8879a758bbe46a6b0b5b9d51a3cd3cb5660106d185b0f1fd981e4475663d08

COPY --from=build /go/src/main /centauri
EXPOSE 8080
EXPOSE 8443
ENTRYPOINT ["/centauri"]
