# Generated from https://github.com/greboid/dockerfiles/blob/master/dockercleanup/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.33.1-r3","apk:ca-certificates":"20191127-r5","apk:expat":"2.4.1-r0","apk:git":"2.32.0-r0","apk:libcrypto1.1":"1.1.1l-r0","apk:libcurl":"7.79.1-r0","apk:libssl1.1":"1.1.1l-r0","apk:musl":"1.2.2-r3","apk:nghttp2-libs":"1.43.0-r0","apk:pcre2":"10.36-r0","apk:zlib":"1.2.11-r3","github:greboid/dockercleanup":"v1.0.0","image:baseroot":"5dea99b6fc9ffe51ff467e5ae1f2ccd6157090f4d20db63d306faa7fe7adac51","image:golang":"b39e962ca9b7c2d31ba231c4912fc7831d59dfbb5dcd5e3fa9bba79bd51cc32c"}

FROM reg.g5d.dev/golang@sha256:b39e962ca9b7c2d31ba231c4912fc7831d59dfbb5dcd5e3fa9bba79bd51cc32c AS build

ARG TAG="v1.0.0"

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
    git clone --depth=1 -b $TAG --single-branch https://github.com/greboid/dockercleanup /go/src/github.com/greboid/dockercleanup; \
    cd /go/src/github.com/greboid/dockercleanup; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    go run github.com/google/go-licenses@latest save ./... --save_path=/notices; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/greboid/dockercleanup/main;

FROM reg.g5d.dev/baseroot@sha256:5dea99b6fc9ffe51ff467e5ae1f2ccd6157090f4d20db63d306faa7fe7adac51

COPY --from=build /go/src/github.com/greboid/dockercleanup/main /dockercleanup
COPY --from=build /notices /notices
EXPOSE 8080
ENTRYPOINT ["/dockercleanup"]
