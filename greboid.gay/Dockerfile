# Generated from https://github.com/csmith/dockerfiles/blob/master/greboid.gay/Dockerfile.gotpl
# BOM: {"apk:bash":"5.1.4-r0","apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.33.1-r3","apk:ca-certificates":"20191127-r5","apk:coreutils":"8.32-r2","apk:expat":"2.4.1-r0","apk:giflib":"5.2.1-r0","apk:git":"2.32.0-r0","apk:libacl":"2.2.53-r0","apk:libattr":"2.5.1-r0","apk:libcrypto1.1":"1.1.1l-r0","apk:libcurl":"7.79.1-r0","apk:libjpeg-turbo":"2.1.0-r0","apk:libpng":"1.6.37-r1","apk:libssl1.1":"1.1.1l-r0","apk:libwebp":"1.2.0-r2","apk:libwebp-tools":"1.2.0-r2","apk:musl":"1.2.2-r3","apk:ncurses-libs":"6.2_p20210612-r0","apk:ncurses-terminfo-base":"6.2_p20210612-r0","apk:nghttp2-libs":"1.43.0-r0","apk:pcre2":"10.36-r0","apk:readline":"8.1.0-r0","apk:s6-ipcserver":"2.10.0.3-r0","apk:skalibs":"2.10.0.3-r0","apk:utmps":"0.1.0.2-r0","apk:zlib":"1.2.11-r3","github:greboid/greboid.gay":"v1.0.0","image:alpine":"sha256:b81c50e56d1dd83921e2dbf786ef233ac11bf297a9130308f5e1da9f549d7ff3","image:base":"sha256:5fc82ed384dc72f0fb8b4c2d57fc45ecdca45f1b4de6fee243464f030864854f","image:golang":"sha256:cc4e1d5e44c7d5f061a9148dab6f33064d8f6a43b2d6e482e9e925285db27244"}

FROM reg.g5d.dev/alpine@sha256:b81c50e56d1dd83921e2dbf786ef233ac11bf297a9130308f5e1da9f549d7ff3 AS images

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
        libcurl=7.79.1-r0 \
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

FROM reg.g5d.dev/golang@sha256:cc4e1d5e44c7d5f061a9148dab6f33064d8f6a43b2d6e482e9e925285db27244 AS build

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
    git clone --depth=1 -b $TAG --single-branch https://github.com/greboid/greboid.gay /go/src/github.com/greboid/greboid.gay; \
    cd /go/src/github.com/greboid/greboid.gay; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/greboid/greboid.gay/main;

FROM reg.g5d.dev/base@sha256:5fc82ed384dc72f0fb8b4c2d57fc45ecdca45f1b4de6fee243464f030864854f

COPY --from=build /go/src/github.com/greboid/greboid.gay/main /greboid.gay
COPY --from=build /go/src/github.com/greboid/greboid.gay/templates /templates
COPY --from=images /app/images /images
EXPOSE 8080
ENTRYPOINT ["/greboid.gay"]
