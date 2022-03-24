# Generated from https://github.com/greboid/dockerfiles/blob/master/hugo/Dockerfile.gotpl
# BOM: {"apk:binutils":"2.37-r3","apk:g++":"10.3.1_git20211027-r0","apk:gcc":"10.3.1_git20211027-r0","apk:gmp":"6.2.1-r1","apk:grep":"3.7-r0","apk:isl22":"0.22-r0","apk:libatomic":"10.3.1_git20211027-r0","apk:libc-dev":"0.7.2-r3","apk:libgcc":"10.3.1_git20211027-r0","apk:libgomp":"10.3.1_git20211027-r0","apk:libgphobos":"10.3.1_git20211027-r0","apk:libstdc++":"10.3.1_git20211027-r0","apk:mpc1":"1.2.1-r0","apk:mpfr4":"4.1.0-r0","apk:musl":"1.2.2-r7","apk:musl-dev":"1.2.2-r7","apk:pcre":"8.45-r1","apk:zlib":"1.2.11-r3","github:gohugoio/hugo":"v0.95.0","image:baseglibc":"7c1ac1ad60a1c9c50ea2390149ea249ec823ae39acfcf22f3ed2e855f885be87","image:golang":"328e0e32945e05d5843e63fb286c7474c089614cc195bd0308255b61f3f8a1c2"}

FROM reg.g5d.dev/golang@sha256:328e0e32945e05d5843e63fb286c7474c089614cc195bd0308255b61f3f8a1c2 AS build

ARG TAG="v0.95.0"

RUN set -eux; \
    apk add --no-cache \
        binutils=2.37-r3 \
        g++=10.3.1_git20211027-r0 \
        gcc=10.3.1_git20211027-r0 \
        gmp=6.2.1-r1 \
        isl22=0.22-r0 \
        libatomic=10.3.1_git20211027-r0 \
        libc-dev=0.7.2-r3 \
        libgcc=10.3.1_git20211027-r0 \
        libgomp=10.3.1_git20211027-r0 \
        libgphobos=10.3.1_git20211027-r0 \
        libstdc++=10.3.1_git20211027-r0 \
        mpc1=1.2.1-r0 \
        mpfr4=4.1.0-r0 \
        musl=1.2.2-r7 \
        musl-dev=1.2.2-r7 \
        zlib=1.2.11-r3 \
        ; \
    git -c advice.detachedHead=false clone --depth=1 -b $TAG --single-branch https://github.com/gohugoio/hugo /src/hugo; \
    cd /src/hugo; \
    CGO_ENABLED=1 go build -a -ldflags "-s" -ldflags "-w" -ldflags "-X github.com/gohugoio/hugo/common/hugo.buildDate=Unknown" -ldflags "-X github.com/gohugoio/hugo/common/hugo.commitHash=$TAG"  -ldflags "-X github.com/gohugoio/hugo/common/hugo.vendorInfo=greboid" -trimpath -tags extended -o main .; \
    mkdir -p /tmp/site;\
    apk add --no-cache \
        binutils=2.37-r3 \
        grep=3.7-r0 \
        libgcc=10.3.1_git20211027-r0 \
        libstdc++=10.3.1_git20211027-r0 \
        musl=1.2.2-r7 \
        pcre=8.45-r1 \
        zlib=1.2.11-r3 \
        ;

FROM reg.g5d.dev/baseglibc@sha256:7c1ac1ad60a1c9c50ea2390149ea249ec823ae39acfcf22f3ed2e855f885be87

ENV HUGO_CACHEDIR="/build/cache"
COPY --from=build /src/hugo/main /usr/bin/hugo
COPY --from=build --chown=65532:65532 /tmp/site /build
WORKDIR /build
CMD ["/usr/bin/hugo", "-s", "/build/site", "-d", "/build/public"]
