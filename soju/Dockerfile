# Generated from https://github.com/greboid/dockerfiles/blob/master/soju/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r14","apk:busybox":"1.36.0-r9","apk:busybox-binsh":"1.36.0-r9","apk:ca-certificates":"20230506-r0","apk:git":"2.40.1-r0","apk:libcrypto3":"3.1.0-r4","apk:libcurl":"8.1.1-r1","apk:libexpat":"2.5.0-r1","apk:libidn2":"2.3.4-r1","apk:libssl3":"3.1.0-r4","apk:libunistring":"1.1-r1","apk:musl":"1.2.4-r0","apk:nghttp2-libs":"1.53.0-r0","apk:patch":"2.7.6-r10","apk:pcre2":"10.42-r1","apk:zlib":"1.2.13-r1","github:csmith/certwrapper":"v4.1.0","github:emersion/soju":"v0.6.1","image:base":"4ddff4468810fe9d9dc3285917d2923c41ae9fb6d1fa2b91c3f1d460cb94427a","image:golang":"6dbc7caa4b67c955ee22d7ccea9caaa9480f9f0fe75dec6dad10fb9a888b9dd7"}

FROM reg.g5d.dev/golang@sha256:6dbc7caa4b67c955ee22d7ccea9caaa9480f9f0fe75dec6dad10fb9a888b9dd7 AS build

ARG SOJUTAG="v0.6.1"
ARG CERTTAG="v4.1.0"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r14 \
        busybox=1.36.0-r9 \
        busybox-binsh=1.36.0-r9 \
        ca-certificates=20230506-r0 \
        git=2.40.1-r0 \
        libcrypto3=3.1.0-r4 \
        libcurl=8.1.1-r1 \
        libexpat=2.5.0-r1 \
        libidn2=2.3.4-r1 \
        libssl3=3.1.0-r4 \
        libunistring=1.1-r1 \
        musl=1.2.4-r0 \
        nghttp2-libs=1.53.0-r0 \
        patch=2.7.6-r10 \
        pcre2=10.42-r1 \
        zlib=1.2.13-r1 \
        ; \
    git clone --depth=1 -b $SOJUTAG --single-branch https://github.com/emersion/soju /go/src/github.com/emersion/soju; \
    cd /go/src/github.com/emersion/soju; \
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o soju ./cmd/soju; \
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o sojuctl ./cmd/sojuctl; \
    git clone --depth=1 -b $CERTTAG --single-branch https://github.com/csmith/certwrapper /go/src/github.com/csmith/certwrapper; \
    cd /go/src/github.com/csmith/certwrapper;\
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -tags lego_httpreq -o certwrapper .; \
    go-licenses save ./... --save_path=/notices; \
    mkdir /data

FROM reg.g5d.dev/base@sha256:4ddff4468810fe9d9dc3285917d2923c41ae9fb6d1fa2b91c3f1d460cb94427a

COPY --from=build /go/src/github.com/emersion/soju/soju /soju
COPY --from=build /go/src/github.com/emersion/soju/sojuctl /sojuctl
COPY --from=build /go/src/github.com/csmith/certwrapper/certwrapper /certwrapper
COPY --from=build /notices /notices
COPY --from=build --chown=65535:65535 /data /data
ENTRYPOINT ["/certwrapper", "/soju"]
