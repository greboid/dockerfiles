# Generated from https://github.com/greboid/dockerfiles/blob/master/soju/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r6","apk:busybox":"1.35.0-r13","apk:ca-certificates":"20211220-r0","apk:expat":"2.4.8-r0","apk:git":"2.36.1-r0","apk:libcrypto1.1":"1.1.1o-r0","apk:libcurl":"7.83.1-r1","apk:libssl1.1":"1.1.1o-r0","apk:musl":"1.2.3-r0","apk:nghttp2-libs":"1.47.0-r0","apk:patch":"2.7.6-r7","apk:pcre2":"10.39-r0","apk:zlib":"1.2.12-r1","github:csmith/certwrapper":"v4.1.0","github:emersion/soju":"v0.4.0","image:base":"301b2ca14deb4f6e198c90f332d09a6b3d00ae5d0d229020f89cd9105589408a","image:golang":"353ced216369febeb109bf954245bae83aff25374ce0d270144b3de95fc1fac8"}

FROM reg.g5d.dev/golang@sha256:353ced216369febeb109bf954245bae83aff25374ce0d270144b3de95fc1fac8 AS build

ARG SOJUTAG="v0.4.0"
ARG CERTTAG="v4.1.0"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r6 \
        busybox=1.35.0-r13 \
        ca-certificates=20211220-r0 \
        expat=2.4.8-r0 \
        git=2.36.1-r0 \
        libcrypto1.1=1.1.1o-r0 \
        libcurl=7.83.1-r1 \
        libssl1.1=1.1.1o-r0 \
        musl=1.2.3-r0 \
        nghttp2-libs=1.47.0-r0 \
        patch=2.7.6-r7 \
        pcre2=10.39-r0 \
        zlib=1.2.12-r1 \
        ; \
    git -c advice.detachedHead=false clone --depth=1 -b $SOJUTAG --single-branch https://github.com/emersion/soju /go/src/github.com/emersion/soju; \
    cd /go/src/github.com/emersion/soju; \
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o soju ./cmd/soju; \
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o sojuctl ./cmd/sojuctl; \
    git -c advice.detachedHead=false clone --depth=1 -b $CERTTAG --single-branch https://github.com/csmith/certwrapper /go/src/github.com/csmith/certwrapper; \
    cd /go/src/github.com/csmith/certwrapper;\
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -tags lego_httpreq -o certwrapper .; \
    go-licenses save ./... --save_path=/notices; \
    mkdir /data

FROM reg.g5d.dev/base@sha256:301b2ca14deb4f6e198c90f332d09a6b3d00ae5d0d229020f89cd9105589408a

COPY --from=build /go/src/github.com/emersion/soju/soju /soju
COPY --from=build /go/src/github.com/emersion/soju/sojuctl /sojuctl
COPY --from=build /go/src/github.com/csmith/certwrapper/certwrapper /certwrapper
COPY --from=build /notices /notices
COPY --from=build --chown=65535:65535 /data /data
ENTRYPOINT ["/certwrapper", "/soju"]
