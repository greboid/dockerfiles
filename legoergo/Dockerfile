# Generated from https://github.com/greboid/dockerfiles/blob/master/legoergo/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.34.1-r5","apk:ca-certificates":"20211220-r0","apk:expat":"2.4.7-r0","apk:git":"2.34.1-r0","apk:libcrypto1.1":"1.1.1n-r0","apk:libcurl":"7.80.0-r0","apk:libssl1.1":"1.1.1n-r0","apk:musl":"1.2.2-r7","apk:nghttp2-libs":"1.46.0-r0","apk:pcre2":"10.39-r0","apk:zlib":"1.2.12-r0","github:csmith/certwrapper":"v4.1.0","github:ergochat/ergo":"v2.9.1","image:base":"ee8879a758bbe46a6b0b5b9d51a3cd3cb5660106d185b0f1fd981e4475663d08","image:golang":"bd8baaef0165a3bc96289502912942c1e9f05c9ee2f152bd7d4e2ba93bc9a1b5"}

FROM reg.g5d.dev/golang@sha256:bd8baaef0165a3bc96289502912942c1e9f05c9ee2f152bd7d4e2ba93bc9a1b5 AS build

ARG CERTTAG="v4.1.0"
ARG ERGOTAG="v2.9.1"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r5 \
        busybox=1.34.1-r5 \
        ca-certificates=20211220-r0 \
        expat=2.4.7-r0 \
        git=2.34.1-r0 \
        libcrypto1.1=1.1.1n-r0 \
        libcurl=7.80.0-r0 \
        libssl1.1=1.1.1n-r0 \
        musl=1.2.2-r7 \
        nghttp2-libs=1.46.0-r0 \
        pcre2=10.39-r0 \
        zlib=1.2.12-r0 \
        ; \
    git -c advice.detachedHead=false clone --depth=1 -b $CERTTAG --single-branch https://github.com/csmith/certwrapper /go/src/github.com/csmith/certwrapper; \
    cd /go/src/github.com/csmith/certwrapper; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -tags lego_httpreq -o main .; \
    git -c advice.detachedHead=false clone --depth=1 -b $ERGOTAG --single-branch https://github.com/ergochat/ergo /go/src/github.com/ergochat/ergo; \
    cd /go/src/github.com/ergochat/ergo; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    go-licenses save ./... --save_path=/notices;

FROM reg.g5d.dev/base@sha256:ee8879a758bbe46a6b0b5b9d51a3cd3cb5660106d185b0f1fd981e4475663d08

COPY --from=build /go/src/github.com/csmith/certwrapper/main /certwrapper
COPY --from=build /go/src/github.com/ergochat/ergo/main /ergo
COPY --from=build /go/src/github.com/ergochat/ergo/languages /ircd-bin/languages
COPY --from=build /notices /notices
CMD ["/certwrapper", "/ergo", "run", "--conf", "/ircd/ircd.yaml"]
