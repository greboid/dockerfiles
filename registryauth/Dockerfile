# Generated from https://github.com/greboid/dockerfiles/blob/master/registryauth/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r2","apk:git":"2.38.1-r0","apk:libcrypto3":"3.0.7-r0","apk:libcurl":"7.86.0-r1","apk:libexpat":"2.5.0-r0","apk:libssl3":"3.0.7-r0","apk:musl":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:pcre2":"10.40-r0","apk:zlib":"1.2.13-r0","github:greboid/registryauth":"v1.0.19","image:base":"9af89525fe357fd63d99e49bba4aa94e6e88a7d436500c3588764112f8862109","image:golang":"5d66ad9a693ec7891f3ee88600d4cde9d58d37d479bcf5422fd9308cd1fa65ee"}

FROM reg.g5d.dev/golang@sha256:5d66ad9a693ec7891f3ee88600d4cde9d58d37d479bcf5422fd9308cd1fa65ee AS build

ARG TAG="v1.0.19"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r9\
        busybox=1.35.0-r29\
        busybox-binsh=1.35.0-r29\
        ca-certificates=20220614-r2\
        git=2.38.1-r0\
        libcrypto3=3.0.7-r0\
        libcurl=7.86.0-r1\
        libexpat=2.5.0-r0\
        libssl3=3.0.7-r0\
        musl=1.2.3-r4\
        nghttp2-libs=1.51.0-r0\
        pcre2=10.40-r0\
        zlib=1.2.13-r0\
        ; \
    git -c advice.detachedHead=false clone --depth=1 -b $TAG --single-branch https://github.com/greboid/registryauth /go/src/github.com/greboid/registryauth; \
    cd /go/src/github.com/greboid/registryauth; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main ./cmd/authedreg; \
    go-licenses save ./... --save_path=/notices; \
    mkdir /data

FROM reg.g5d.dev/base@sha256:9af89525fe357fd63d99e49bba4aa94e6e88a7d436500c3588764112f8862109

COPY --from=build /go/src/github.com/greboid/registryauth/main /authedreg
COPY --from=build --chown=65532:65532 /data /data
COPY --from=build /notices /notices
EXPOSE 8080
ENTRYPOINT ["/authedreg"]
