# Generated from https://github.com/csmith/dockerfiles/blob/master/greboid.gay/Dockerfile.gotpl
# BOM: {"apk:bash":"5.1.4-r0","apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.33.1-r3","apk:ca-certificates":"20191127-r5","apk:coreutils":"8.32-r2","apk:expat":"2.4.1-r0","apk:giflib":"5.2.1-r0","apk:git":"2.32.0-r0","apk:libacl":"2.2.53-r0","apk:libattr":"2.5.1-r0","apk:libcrypto1.1":"1.1.1l-r0","apk:libcurl":"7.79.0-r0","apk:libjpeg-turbo":"2.1.0-r0","apk:libpng":"1.6.37-r1","apk:libssl1.1":"1.1.1l-r0","apk:libwebp":"1.2.0-r2","apk:libwebp-tools":"1.2.0-r2","apk:musl":"1.2.2-r3","apk:ncurses-libs":"6.2_p20210612-r0","apk:ncurses-terminfo-base":"6.2_p20210612-r0","apk:nghttp2-libs":"1.43.0-r0","apk:pcre2":"10.36-r0","apk:readline":"8.1.0-r0","apk:s6-ipcserver":"2.10.0.3-r0","apk:skalibs":"2.10.0.3-r0","apk:utmps":"0.1.0.2-r0","apk:zlib":"1.2.11-r3","github:greboid/greboid.gay":"v1.0.0","image:alpine":"sha256:f32b56d2f268e9a4dce68b95adac9f81b8ed8f9bc7fd33ba0d804e7426f603fd","image:base":"sha256:9eb1693b711c4282b96c772aecfb3fe764a384ed907fb24f80831830e90678c8","image:golang":"sha256:c1d9d794a310682b45fbdaad1dbe94ee3773ce218c3c0fcdff6b499cc87c6f4c"}

FROM reg.g5d.dev/alpine@sha256:f32b56d2f268e9a4dce68b95adac9f81b8ed8f9bc7fd33ba0d804e7426f603fd AS images

ARG TAG="v1.0.0"

RUN set -eux; \
    # Dependencies and source for bootstrapping go
    apk add --no-cache --virtual .build-deps \
        bash=5.1.4-r0 \
        brotli-libs=1.0.9-r5 \
        busybox=1.33.1-r3 \
        ca-certificates=20191127-r5 \
        coreutils=8.32-r2 \
        expat=2.4.1-r0 \
        giflib=5.2.1-r0 \
        git=2.32.0-r0 \
        libacl=2.2.53-r0 \
        libattr=2.5.1-r0 \
        libcrypto1.1=1.1.1l-r0 \
        libcurl=7.79.0-r0 \
        libjpeg-turbo=2.1.0-r0 \
        libpng=1.6.37-r1 \
        libssl1.1=1.1.1l-r0 \
        libwebp=1.2.0-r2 \
        libwebp-tools=1.2.0-r2 \
        musl=1.2.2-r3 \
        ncurses-libs=6.2_p20210612-r0 \
        ncurses-terminfo-base=6.2_p20210612-r0 \
        nghttp2-libs=1.43.0-r0 \
        pcre2=10.36-r0 \
        readline=8.1.0-r0 \
        s6-ipcserver=2.10.0.3-r0 \
        skalibs=2.10.0.3-r0 \
        utmps=0.1.0.2-r0 \
        zlib=1.2.11-r3 \
        ; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/greboid/greboid.gay /go/src/github.com/greboid/greboid.gay; \
    cd /go/src/github.com/greboid/greboid.gay; \
    mkdir -p /app/images; \
    cp -r images /app; \
    chmod +x minify.sh; \
    ./minify.sh

FROM reg.g5d.dev/golang@sha256:c1d9d794a310682b45fbdaad1dbe94ee3773ce218c3c0fcdff6b499cc87c6f4c AS build

ARG TAG="v1.0.0"

RUN set -eux; \
    apk add --no-cache git; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/greboid/greboid.gay /go/src/github.com/greboid/greboid.gay; \
    cd /go/src/github.com/greboid/greboid.gay; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/greboid/greboid.gay/main;

FROM reg.g5d.dev/base@sha256:9eb1693b711c4282b96c772aecfb3fe764a384ed907fb24f80831830e90678c8

COPY --from=build /go/src/github.com/greboid/greboid.gay/main /greboid.gay
COPY --from=build /go/src/github.com/greboid/greboid.gay/templates /templates
COPY --from=images /app/images /images
EXPOSE 8080
ENTRYPOINT ["/greboid.gay"]
