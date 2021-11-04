# Generated from https://github.com/greboid/dockerfiles/blob/master/linx-server/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.33.1-r3","apk:ca-certificates":"20191127-r5","apk:expat":"2.4.1-r0","apk:git":"2.32.0-r0","apk:libcrypto1.1":"1.1.1l-r0","apk:libcurl":"7.79.1-r0","apk:libssl1.1":"1.1.1l-r0","apk:musl":"1.2.2-r3","apk:nghttp2-libs":"1.43.0-r0","apk:pcre2":"10.36-r0","apk:zlib":"1.2.11-r3","github:csmith/linx-server":"v2.3.9","image:base":"f25b9dc29d96dee9ff3919d93729ba9105802a4ee8a15b69e654877fa74e0af2","image:golang":"b39e962ca9b7c2d31ba231c4912fc7831d59dfbb5dcd5e3fa9bba79bd51cc32c"}

FROM reg.g5d.dev/golang@sha256:b39e962ca9b7c2d31ba231c4912fc7831d59dfbb5dcd5e3fa9bba79bd51cc32c AS build

ARG TAG="v2.3.9"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r5\
        busybox=1.33.1-r3\
        ca-certificates=20191127-r5\
        expat=2.4.1-r0\
        git=2.32.0-r0\
        libcrypto1.1=1.1.1l-r0\
        libcurl=7.79.1-r0\
        libssl1.1=1.1.1l-r0\
        musl=1.2.2-r3\
        nghttp2-libs=1.43.0-r0\
        pcre2=10.36-r0\
        zlib=1.2.11-r3\
        ; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/csmith/linx-server /go/src/github.com/csmith/linx-server; \
    cd /go/src/github.com/csmith/linx-server; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    go run github.com/google/go-licenses@latest save ./... --save_path=/notices; \
    mkdir -p /data/files /data/meta; \
    chown -R 65532:65532 /data; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/csmith/linx-server/main;

FROM reg.g5d.dev/base@sha256:f25b9dc29d96dee9ff3919d93729ba9105802a4ee8a15b69e654877fa74e0af2

COPY --from=build /go/src/github.com/csmith/linx-server/main /linx-server
COPY --from=build /data /data
COPY --from=build /notices /notices

VOLUME ["/data/files", "/data/meta"]
EXPOSE 8080
ENTRYPOINT ["/linx-server", "-bind=0.0.0.0:8080", "-filespath=/data/files/", "-metapath=/data/meta/"]
CMD ["-sitename=linx", "-allowhotlink"]
