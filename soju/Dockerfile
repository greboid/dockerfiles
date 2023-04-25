# Generated from https://github.com/greboid/dockerfiles/blob/master/soju/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r4","apk:git":"2.38.4-r1","apk:libcrypto3":"3.0.8-r4","apk:libcurl":"8.0.1-r0","apk:libexpat":"2.5.0-r0","apk:libssl3":"3.0.8-r4","apk:musl":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:patch":"2.7.6-r9","apk:pcre2":"10.42-r0","apk:zlib":"1.2.13-r0","github:csmith/certwrapper":"v4.1.0","github:emersion/soju":"v0.6.1","image:base":"ef8b65fda5d776551a87f970b02ebd1aa67f98c60b68637d910429a4e18be9b9","image:golang":"5d04510065ded5dc194b5809e8f9acc159e9cc6f29b239ef69fb11c92a528d0f"}

FROM reg.g5d.dev/golang@sha256:5d04510065ded5dc194b5809e8f9acc159e9cc6f29b239ef69fb11c92a528d0f AS build

ARG SOJUTAG="v0.6.1"
ARG CERTTAG="v4.1.0"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r9 \
        busybox=1.35.0-r29 \
        busybox-binsh=1.35.0-r29 \
        ca-certificates=20220614-r4 \
        git=2.38.4-r1 \
        libcrypto3=3.0.8-r4 \
        libcurl=8.0.1-r0 \
        libexpat=2.5.0-r0 \
        libssl3=3.0.8-r4 \
        musl=1.2.3-r4 \
        nghttp2-libs=1.51.0-r0 \
        patch=2.7.6-r9 \
        pcre2=10.42-r0 \
        zlib=1.2.13-r0 \
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

FROM reg.g5d.dev/base@sha256:ef8b65fda5d776551a87f970b02ebd1aa67f98c60b68637d910429a4e18be9b9

COPY --from=build /go/src/github.com/emersion/soju/soju /soju
COPY --from=build /go/src/github.com/emersion/soju/sojuctl /sojuctl
COPY --from=build /go/src/github.com/csmith/certwrapper/certwrapper /certwrapper
COPY --from=build /notices /notices
COPY --from=build --chown=65535:65535 /data /data
ENTRYPOINT ["/certwrapper", "/soju"]
