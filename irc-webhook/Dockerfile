# Generated from https://github.com/greboid/dockerfiles/blob/master/irc-webhook/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r6","apk:busybox":"1.35.0-r14","apk:ca-certificates":"20211220-r0","apk:expat":"2.4.8-r0","apk:git":"2.36.1-r0","apk:libcrypto1.1":"1.1.1p-r0","apk:libcurl":"7.83.1-r2","apk:libssl1.1":"1.1.1p-r0","apk:musl":"1.2.3-r0","apk:nghttp2-libs":"1.47.0-r0","apk:pcre2":"10.40-r0","apk:zlib":"1.2.12-r1","github:greboid/irc-webhook":"v4.0.1","image:base":"d8e173562b2af5c2d2a6bedc5af4a36a9fe9e9561aca25d22c947752a1695a64","image:golang":"8a7ef4bb7eec3e9eef301d586b5517c54b9c6fe1e2c0d06f74f0c0d7b71b62e5"}

FROM reg.g5d.dev/golang@sha256:8a7ef4bb7eec3e9eef301d586b5517c54b9c6fe1e2c0d06f74f0c0d7b71b62e5 AS build

ARG TAG="v4.0.1"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r6\
        busybox=1.35.0-r14\
        ca-certificates=20211220-r0\
        expat=2.4.8-r0\
        git=2.36.1-r0\
        libcrypto1.1=1.1.1p-r0\
        libcurl=7.83.1-r2\
        libssl1.1=1.1.1p-r0\
        musl=1.2.3-r0\
        nghttp2-libs=1.47.0-r0\
        pcre2=10.40-r0\
        zlib=1.2.12-r1\
        ; \
    git -c advice.detachedHead=false clone --depth=1 -b $TAG --single-branch https://github.com/greboid/irc-webhook /go/src/github.com/greboid/irc-webhook; \
    cd /go/src/github.com/greboid/irc-webhook; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main ./cmd/webhook; \
    go-licenses save ./... --save_path=/notices;

FROM reg.g5d.dev/base@sha256:d8e173562b2af5c2d2a6bedc5af4a36a9fe9e9561aca25d22c947752a1695a64

COPY --from=build /go/src/github.com/greboid/irc-webhook/main /irc-webhook
COPY --from=build /notices /notices
ENTRYPOINT ["/irc-webhook"]
