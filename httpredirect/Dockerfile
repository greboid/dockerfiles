# Generated from https://github.com/greboid/dockerfiles/blob/master/httpredirect/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.34.1-r3","apk:ca-certificates":"20191127-r7","apk:expat":"2.4.1-r0","apk:git":"2.34.1-r0","apk:libcrypto1.1":"1.1.1l-r7","apk:libcurl":"7.80.0-r0","apk:libssl1.1":"1.1.1l-r7","apk:musl":"1.2.2-r7","apk:nghttp2-libs":"1.46.0-r0","apk:pcre2":"10.39-r0","apk:zlib":"1.2.11-r3","github:greboid/httpredirect":"v1.0.0","image:base":"e0de705f0d65b2dab86345626fa5ef6279e04bfbfe189ad9dfa3c5d678428ed9","image:golang":"adbc935862ae886bd37a2f6ad02e6e227c1c1ca5e592fd2300970548acce783e"}

FROM reg.g5d.dev/golang@sha256:adbc935862ae886bd37a2f6ad02e6e227c1c1ca5e592fd2300970548acce783e AS build

ARG TAG="v1.0.0"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r5\
        busybox=1.34.1-r3\
        ca-certificates=20191127-r7\
        expat=2.4.1-r0\
        git=2.34.1-r0\
        libcrypto1.1=1.1.1l-r7\
        libcurl=7.80.0-r0\
        libssl1.1=1.1.1l-r7\
        musl=1.2.2-r7\
        nghttp2-libs=1.46.0-r0\
        pcre2=10.39-r0\
        zlib=1.2.11-r3\
        ; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/greboid/httpredirect /go/src/github.com/greboid/httpredirect; \
    cd /go/src/github.com/greboid/httpredirect; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    go run github.com/google/go-licenses@latest save ./... --save_path=/notices; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/greboid/httpredirect/main;

FROM reg.g5d.dev/base@sha256:e0de705f0d65b2dab86345626fa5ef6279e04bfbfe189ad9dfa3c5d678428ed9

COPY --from=build /go/src/github.com/greboid/httpredirect/main /httpredirect
COPY --from=build /notices /notices
EXPOSE 8080
ENTRYPOINT ["/httpredirect"]
