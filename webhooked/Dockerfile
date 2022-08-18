# Generated from https://github.com/greboid/dockerfiles/blob/master/webhooked/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r6","apk:busybox":"1.35.0-r17","apk:ca-certificates":"20220614-r0","apk:expat":"2.4.8-r0","apk:git":"2.36.2-r0","apk:libcrypto1.1":"1.1.1q-r0","apk:libcurl":"7.83.1-r2","apk:libssl1.1":"1.1.1q-r0","apk:musl":"1.2.3-r0","apk:nghttp2-libs":"1.47.0-r0","apk:pcre2":"10.40-r0","apk:zlib":"1.2.12-r3","github:csmith/webhooked":"v1.0.0","image:base":"148dcee1d765d04d178e78cbf9e1c44b4c23af1097d0e79418893051e0e13c69","image:golang":"2bf2ea4e192e0a259cfe76a159679ff837fc2ac3c154899ef0ecb485c57c0ffe"}

FROM reg.g5d.dev/golang@sha256:2bf2ea4e192e0a259cfe76a159679ff837fc2ac3c154899ef0ecb485c57c0ffe AS build

ARG TAG="v1.0.0"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r6\
        busybox=1.35.0-r17\
        ca-certificates=20220614-r0\
        expat=2.4.8-r0\
        git=2.36.2-r0\
        libcrypto1.1=1.1.1q-r0\
        libcurl=7.83.1-r2\
        libssl1.1=1.1.1q-r0\
        musl=1.2.3-r0\
        nghttp2-libs=1.47.0-r0\
        pcre2=10.40-r0\
        zlib=1.2.12-r3\
        ; \
    git -c advice.detachedHead=false clone --depth=1 -b $TAG --single-branch https://github.com/csmith/webhooked /go/src/github.com/csmith/webhooked; \
    cd /go/src/github.com/csmith/webhooked; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    go-licenses save ./... --save_path=/notices;

FROM reg.g5d.dev/base@sha256:148dcee1d765d04d178e78cbf9e1c44b4c23af1097d0e79418893051e0e13c69

COPY --from=build /go/src/github.com/csmith/webhooked/main /webhooked
COPY --from=build /notices /notices

ENTRYPOINT ["/webhooked"]
