# Generated from https://github.com/greboid/dockerfiles/blob/master/greboid.com/Dockerfile.gotpl
# BOM: {"apk:bash":"5.1.8-r0","apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.34.1-r3","apk:ca-certificates":"20211220-r0","apk:coreutils":"9.0-r2","apk:expat":"2.4.3-r0","apk:giflib":"5.2.1-r0","apk:git":"2.34.1-r0","apk:libacl":"2.2.53-r0","apk:libattr":"2.5.1-r1","apk:libcrypto1.1":"1.1.1l-r8","apk:libcurl":"7.80.0-r0","apk:libjpeg-turbo":"2.1.2-r0","apk:libpng":"1.6.37-r1","apk:libssl1.1":"1.1.1l-r8","apk:libwebp":"1.2.2-r0","apk:libwebp-tools":"1.2.2-r0","apk:musl":"1.2.2-r7","apk:ncurses-libs":"6.3_p20211120-r0","apk:ncurses-terminfo-base":"6.3_p20211120-r0","apk:nghttp2-libs":"1.46.0-r0","apk:pcre2":"10.39-r0","apk:readline":"8.1.1-r0","apk:s6-ipcserver":"2.11.0.0-r0","apk:skalibs":"2.11.0.0-r0","apk:utmps":"0.1.0.3-r0","apk:zlib":"1.2.11-r3","github:greboid/greboid.com":"v1.0.1","image:alpine":"ef95559f475cdf56e55db71c168f1616d497e2215bbd9f6bc63398036ae546a1","image:base":"e390636a8773e2b84066cdbde990e80b3c97b8e7e25428df9f4b8f097323f861","image:golang":"ce2e890f1d9bc1ba9dce4271d1c3e17031d896af1bbc53ffbba1295dea19f2e6"}

FROM reg.g5d.dev/alpine@sha256:ef95559f475cdf56e55db71c168f1616d497e2215bbd9f6bc63398036ae546a1 AS images

ARG TAG="v1.0.1"

RUN set -eux; \
    # Dependencies and source for bootstrapping go
    apk add --no-cache --virtual .build-deps \
        bash=5.1.8-r0 \
        brotli-libs=1.0.9-r5 \
        busybox=1.34.1-r3 \
        ca-certificates=20211220-r0 \
        coreutils=9.0-r2 \
        expat=2.4.3-r0 \
        giflib=5.2.1-r0 \
        git=2.34.1-r0 \
        libacl=2.2.53-r0 \
        libattr=2.5.1-r1 \
        libcrypto1.1=1.1.1l-r8 \
        libcurl=7.80.0-r0 \
        libjpeg-turbo=2.1.2-r0 \
        libpng=1.6.37-r1 \
        libssl1.1=1.1.1l-r8 \
        libwebp=1.2.2-r0 \
        libwebp-tools=1.2.2-r0 \
        musl=1.2.2-r7 \
        ncurses-libs=6.3_p20211120-r0 \
        ncurses-terminfo-base=6.3_p20211120-r0 \
        nghttp2-libs=1.46.0-r0 \
        pcre2=10.39-r0 \
        readline=8.1.1-r0 \
        s6-ipcserver=2.11.0.0-r0 \
        skalibs=2.11.0.0-r0 \
        utmps=0.1.0.3-r0 \
        zlib=1.2.11-r3 \
        ; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/greboid/greboid.com /go/src/github.com/greboid/greboid.com; \
    cd /go/src/github.com/greboid/greboid.com; \
    mkdir -p /app/static/images; \
    cp -r static/images /app/static; \
    chmod +x minify.sh; \
    ./minify.sh

FROM ghcr.io/greboid/cv:latest as cv

FROM reg.g5d.dev/golang@sha256:ce2e890f1d9bc1ba9dce4271d1c3e17031d896af1bbc53ffbba1295dea19f2e6 AS build

ARG TAG="v1.0.1"

COPY --from=images /app/static/images /images
COPY --from=cv /cv.pdf /cv.pdf

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r5\
        busybox=1.34.1-r3\
        ca-certificates=20211220-r0\
        expat=2.4.3-r0\
        git=2.34.1-r0\
        libcrypto1.1=1.1.1l-r8\
        libcurl=7.80.0-r0\
        libssl1.1=1.1.1l-r8\
        musl=1.2.2-r7\
        nghttp2-libs=1.46.0-r0\
        pcre2=10.39-r0\
        zlib=1.2.11-r3\
        ; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/greboid/greboid.com /go/src/github.com/greboid/greboid.com; \
    cd /go/src/github.com/greboid/greboid.com; \
    cp -r /images/* static/images/; \
    cp /cv.pdf static/cv.pdf; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/greboid/greboid.com/main;

FROM reg.g5d.dev/base@sha256:e390636a8773e2b84066cdbde990e80b3c97b8e7e25428df9f4b8f097323f861

COPY --from=build /go/src/github.com/greboid/greboid.com/main /greboid.com
EXPOSE 8080
ENTRYPOINT ["/greboid.com"]
