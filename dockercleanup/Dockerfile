# Generated from https://github.com/greboid/dockerfiles/blob/master/dockercleanup/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r6","apk:busybox":"1.35.0-r14","apk:ca-certificates":"20211220-r0","apk:expat":"2.4.8-r0","apk:git":"2.36.1-r0","apk:libcrypto1.1":"1.1.1o-r0","apk:libcurl":"7.83.1-r1","apk:libssl1.1":"1.1.1o-r0","apk:musl":"1.2.3-r0","apk:nghttp2-libs":"1.47.0-r0","apk:pcre2":"10.40-r0","apk:zlib":"1.2.12-r1","github:greboid/dockercleanup":"v1.0.0","image:base":"d8e173562b2af5c2d2a6bedc5af4a36a9fe9e9561aca25d22c947752a1695a64","image:golang":"7a22f9869b56192b2e2714e0d173b3616733ef5dd5187079d4feee804903172c"}

FROM reg.g5d.dev/golang@sha256:7a22f9869b56192b2e2714e0d173b3616733ef5dd5187079d4feee804903172c AS build

ARG TAG="v1.0.0"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r6\
        busybox=1.35.0-r14\
        ca-certificates=20211220-r0\
        expat=2.4.8-r0\
        git=2.36.1-r0\
        libcrypto1.1=1.1.1o-r0\
        libcurl=7.83.1-r1\
        libssl1.1=1.1.1o-r0\
        musl=1.2.3-r0\
        nghttp2-libs=1.47.0-r0\
        pcre2=10.40-r0\
        zlib=1.2.12-r1\
        ; \
    git -c advice.detachedHead=false clone --depth=1 -b $TAG --single-branch https://github.com/greboid/dockercleanup /go/src/github.com/greboid/dockercleanup; \
    cd /go/src/github.com/greboid/dockercleanup; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    go-licenses save ./... --save_path=/notices;

FROM reg.g5d.dev/base@sha256:d8e173562b2af5c2d2a6bedc5af4a36a9fe9e9561aca25d22c947752a1695a64

COPY --from=build /go/src/github.com/greboid/dockercleanup/main /dockercleanup
COPY --from=build /notices /notices
EXPOSE 8080
ENTRYPOINT ["/dockercleanup"]
