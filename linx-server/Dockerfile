# Generated from https://github.com/greboid/dockerfiles/blob/master/linx-server/Containerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r14","apk:busybox":"1.36.1-r1","apk:busybox-binsh":"1.36.1-r1","apk:ca-certificates":"20230506-r0","apk:git":"2.40.1-r0","apk:libcrypto3":"3.1.1-r3","apk:libcurl":"8.2.0-r1","apk:libexpat":"2.5.0-r1","apk:libidn2":"2.3.4-r1","apk:libssl3":"3.1.1-r3","apk:libunistring":"1.1-r1","apk:musl":"1.2.4-r0","apk:nghttp2-libs":"1.55.1-r0","apk:pcre2":"10.42-r1","apk:zlib":"1.2.13-r1","github:csmith/linx-server":"v3.0.0","image:base":"5c298191cd12c92b8b3df64ab907f33118d5ddffa99f0c31d43c9623330c46e2","image:golang":"36190d5ec7bcb2ccd44dcffdcc6376553799520817f7654c1cf63883a285695f"}

FROM reg.g5d.dev/golang@sha256:36190d5ec7bcb2ccd44dcffdcc6376553799520817f7654c1cf63883a285695f AS build

ARG TAG="v3.0.0"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r14\
        busybox=1.36.1-r1\
        busybox-binsh=1.36.1-r1\
        ca-certificates=20230506-r0\
        git=2.40.1-r0\
        libcrypto3=3.1.1-r3\
        libcurl=8.2.0-r1\
        libexpat=2.5.0-r1\
        libidn2=2.3.4-r1\
        libssl3=3.1.1-r3\
        libunistring=1.1-r1\
        musl=1.2.4-r0\
        nghttp2-libs=1.55.1-r0\
        pcre2=10.42-r1\
        zlib=1.2.13-r1\
        ; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/csmith/linx-server /go/src/github.com/csmith/linx-server; \
    cd /go/src/github.com/csmith/linx-server; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    go-licenses save ./... --save_path=/notices; \
    mkdir -p /data/files /data/meta; \
    chown -R 65532:65532 /data;

FROM reg.g5d.dev/base@sha256:5c298191cd12c92b8b3df64ab907f33118d5ddffa99f0c31d43c9623330c46e2

COPY --from=build /go/src/github.com/csmith/linx-server/main /linx-server
COPY --from=build /data /data
COPY --from=build /notices /notices

VOLUME ["/data/files", "/data/meta"]
EXPOSE 8080
ENTRYPOINT ["/linx-server", "-bind=0.0.0.0:8080", "-filespath=/data/files/", "-metapath=/data/meta/"]
CMD ["-sitename=linx", "-allowhotlink"]
