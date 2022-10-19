# Generated from https://github.com/greboid/dockerfiles/blob/master/goplum/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r6","apk:busybox":"1.35.0-r17","apk:ca-certificates":"20220614-r0","apk:expat":"2.4.9-r0","apk:git":"2.36.3-r0","apk:libcrypto1.1":"1.1.1q-r0","apk:libcurl":"7.83.1-r3","apk:libssl1.1":"1.1.1q-r0","apk:musl":"1.2.3-r0","apk:nghttp2-libs":"1.47.0-r0","apk:pcre2":"10.40-r0","apk:zlib":"1.2.12-r3","github:csmith/goplum":"v0.6.0","image:base":"148dcee1d765d04d178e78cbf9e1c44b4c23af1097d0e79418893051e0e13c69","image:golang":"15a9960c27d67316b3ba3da4c8c7a588e57ae6a7c811f6edda91aced6bb5f573"}

FROM reg.g5d.dev/golang@sha256:15a9960c27d67316b3ba3da4c8c7a588e57ae6a7c811f6edda91aced6bb5f573 AS build

ARG TAG="v0.6.0"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r6\
        busybox=1.35.0-r17\
        ca-certificates=20220614-r0\
        expat=2.4.9-r0\
        git=2.36.3-r0\
        libcrypto1.1=1.1.1q-r0\
        libcurl=7.83.1-r3\
        libssl1.1=1.1.1q-r0\
        musl=1.2.3-r0\
        nghttp2-libs=1.47.0-r0\
        pcre2=10.40-r0\
        zlib=1.2.12-r3\
        ; \
    git -c advice.detachedHead=false clone --depth=1 -b $TAG --single-branch https://github.com/csmith/goplum /go/src/github.com/csmith/goplum; \
    cd /go/src/github.com/csmith/goplum; \
    #Build the plugins (all, its easier)
    for plugin in $(ls plugins); do go build -o $plugin.so -buildmode=plugin ./plugins/$plugin/cmd; done; \
    #Build goplum
    go build -o main ./cmd/goplum; \
    go-licenses save ./... --save_path=/notices;

FROM reg.g5d.dev/base@sha256:148dcee1d765d04d178e78cbf9e1c44b4c23af1097d0e79418893051e0e13c69

COPY --from=build /go/src/github.com/csmith/goplum/main /goplum
COPY --from=build /go/src/github.com/csmith/goplum/*.so /plugins/
COPY --from=build /notices /notices
EXPOSE 8080
ENTRYPOINT ["/goplum"]
