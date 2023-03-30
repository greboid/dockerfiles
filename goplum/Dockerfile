# Generated from https://github.com/greboid/dockerfiles/blob/master/goplum/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r4","apk:git":"2.38.4-r1","apk:libcrypto3":"3.0.8-r3","apk:libcurl":"7.88.1-r1","apk:libexpat":"2.5.0-r0","apk:libssl3":"3.0.8-r3","apk:musl":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:pcre2":"10.42-r0","apk:zlib":"1.2.13-r0","github:csmith/goplum":"v0.6.0","image:base":"ef8b65fda5d776551a87f970b02ebd1aa67f98c60b68637d910429a4e18be9b9","image:golang":"c8e2e67383482171eb5457b17bcd759e11e839c41165c45d8222e53c91bb57e7"}

FROM reg.g5d.dev/golang@sha256:c8e2e67383482171eb5457b17bcd759e11e839c41165c45d8222e53c91bb57e7 AS build

ARG TAG="v0.6.0"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r9\
        busybox=1.35.0-r29\
        busybox-binsh=1.35.0-r29\
        ca-certificates=20220614-r4\
        git=2.38.4-r1\
        libcrypto3=3.0.8-r3\
        libcurl=7.88.1-r1\
        libexpat=2.5.0-r0\
        libssl3=3.0.8-r3\
        musl=1.2.3-r4\
        nghttp2-libs=1.51.0-r0\
        pcre2=10.42-r0\
        zlib=1.2.13-r0\
        ; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/csmith/goplum /go/src/github.com/csmith/goplum; \
    cd /go/src/github.com/csmith/goplum; \
    #Build the plugins (all, its easier)
    for plugin in $(ls plugins); do go build -o $plugin.so -buildmode=plugin ./plugins/$plugin/cmd; done; \
    #Build goplum
    go build -o main ./cmd/goplum; \
    go-licenses save ./... --save_path=/notices;

FROM reg.g5d.dev/base@sha256:ef8b65fda5d776551a87f970b02ebd1aa67f98c60b68637d910429a4e18be9b9

COPY --from=build /go/src/github.com/csmith/goplum/main /goplum
COPY --from=build /go/src/github.com/csmith/goplum/*.so /plugins/
COPY --from=build /notices /notices
EXPOSE 8080
ENTRYPOINT ["/goplum"]
