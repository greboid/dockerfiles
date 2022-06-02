# Generated from https://github.com/greboid/dockerfiles/blob/master/hugo/Dockerfile.gotpl
# BOM: {"apk:binutils":"2.38-r2","apk:g++":"11.2.1_git20220219-r2","apk:gcc":"11.2.1_git20220219-r2","apk:gmp":"6.2.1-r2","apk:grep":"3.7-r0","apk:isl22":"0.22-r0","apk:libatomic":"11.2.1_git20220219-r2","apk:libc-dev":"0.7.2-r3","apk:libgcc":"11.2.1_git20220219-r2","apk:libgomp":"11.2.1_git20220219-r2","apk:libstdc++":"11.2.1_git20220219-r2","apk:mpc1":"1.2.1-r0","apk:mpfr4":"4.1.0-r0","apk:musl":"1.2.3-r0","apk:musl-dev":"1.2.3-r0","apk:pcre":"8.45-r2","apk:zlib":"1.2.12-r1","github:gohugoio/hugo":"v0.100.1","image:baseglibc":"7fded4d2da8681eb619519b4d402500a9f671c2e4fe2917e9f4b5de00b564bce","image:golang":"353ced216369febeb109bf954245bae83aff25374ce0d270144b3de95fc1fac8"}

FROM reg.g5d.dev/golang@sha256:353ced216369febeb109bf954245bae83aff25374ce0d270144b3de95fc1fac8 AS build

ARG TAG="v0.100.1"

RUN set -eux; \
    apk add --no-cache \
        binutils=2.38-r2 \
        g++=11.2.1_git20220219-r2 \
        gcc=11.2.1_git20220219-r2 \
        gmp=6.2.1-r2 \
        isl22=0.22-r0 \
        libatomic=11.2.1_git20220219-r2 \
        libc-dev=0.7.2-r3 \
        libgcc=11.2.1_git20220219-r2 \
        libgomp=11.2.1_git20220219-r2 \
        libstdc++=11.2.1_git20220219-r2 \
        mpc1=1.2.1-r0 \
        mpfr4=4.1.0-r0 \
        musl=1.2.3-r0 \
        musl-dev=1.2.3-r0 \
        zlib=1.2.12-r1 \
        ; \
    git -c advice.detachedHead=false clone --depth=1 -b $TAG --single-branch https://github.com/gohugoio/hugo /src/hugo; \
    cd /src/hugo; \
    CGO_ENABLED=1 go build -a -ldflags "-s" -ldflags "-w" -ldflags "-X github.com/gohugoio/hugo/common/hugo.buildDate=Unknown" -ldflags "-X github.com/gohugoio/hugo/common/hugo.commitHash=$TAG"  -ldflags "-X github.com/gohugoio/hugo/common/hugo.vendorInfo=greboid" -trimpath -tags extended -o main .; \
    mkdir -p /tmp/site;\
    apk add --no-cache \
        binutils=2.38-r2 \
        grep=3.7-r0 \
        libgcc=11.2.1_git20220219-r2 \
        libstdc++=11.2.1_git20220219-r2 \
        musl=1.2.3-r0 \
        pcre=8.45-r2 \
        zlib=1.2.12-r1 \
        ;

FROM reg.g5d.dev/baseglibc@sha256:7fded4d2da8681eb619519b4d402500a9f671c2e4fe2917e9f4b5de00b564bce

ENV HUGO_CACHEDIR="/build/cache"
COPY --from=build /src/hugo/main /usr/bin/hugo
COPY --from=build --chown=65532:65532 /tmp/site /build
WORKDIR /build
CMD ["/usr/bin/hugo", "-s", "/build/site", "-d", "/build/public"]
