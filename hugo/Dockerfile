# Generated from https://github.com/greboid/dockerfiles/blob/master/hugo/Dockerfile.gotpl
# BOM: {"apk:binutils":"2.39-r2","apk:g++":"12.2.1_git20220924-r4","apk:gcc":"12.2.1_git20220924-r4","apk:gmp":"6.2.1-r2","apk:grep":"3.8-r1","apk:isl24":"0.24-r1","apk:libatomic":"12.2.1_git20220924-r4","apk:libc-dev":"0.7.2-r3","apk:libgcc":"12.2.1_git20220924-r4","apk:libgomp":"12.2.1_git20220924-r4","apk:libstdc++":"12.2.1_git20220924-r4","apk:libstdc++-dev":"12.2.1_git20220924-r4","apk:mpc1":"1.2.1-r1","apk:mpfr4":"4.1.0-r0","apk:musl":"1.2.3-r4","apk:musl-dev":"1.2.3-r4","apk:pcre2":"10.42-r0","apk:zlib":"1.2.13-r0","github:gohugoio/hugo":"v0.108.0","image:baseglibc":"5400aeaae743c557874709c0a0c5208e77d89582014989e51a13e7cae6b429e4","image:golang":"a28659b5bb93810642775030406b20b18aa645416d66b8c0f265f02cffac0925"}

FROM reg.g5d.dev/golang@sha256:a28659b5bb93810642775030406b20b18aa645416d66b8c0f265f02cffac0925 AS build

ARG TAG="v0.108.0"

RUN set -eux; \
    apk add --no-cache \
        binutils=2.39-r2 \
        g++=12.2.1_git20220924-r4 \
        gcc=12.2.1_git20220924-r4 \
        gmp=6.2.1-r2 \
        isl24=0.24-r1 \
        libatomic=12.2.1_git20220924-r4 \
        libc-dev=0.7.2-r3 \
        libgcc=12.2.1_git20220924-r4 \
        libgomp=12.2.1_git20220924-r4 \
        libstdc++=12.2.1_git20220924-r4 \
        libstdc++-dev=12.2.1_git20220924-r4 \
        mpc1=1.2.1-r1 \
        mpfr4=4.1.0-r0 \
        musl=1.2.3-r4 \
        musl-dev=1.2.3-r4 \
        zlib=1.2.13-r0 \
        ; \
    git -c advice.detachedHead=false clone --depth=1 -b $TAG --single-branch https://github.com/gohugoio/hugo /src/hugo; \
    cd /src/hugo; \
    CGO_ENABLED=1 go build -a -ldflags "-s" -ldflags "-w" -ldflags "-X github.com/gohugoio/hugo/common/hugo.buildDate=Unknown" -ldflags "-X github.com/gohugoio/hugo/common/hugo.commitHash=$TAG"  -ldflags "-X github.com/gohugoio/hugo/common/hugo.vendorInfo=greboid" -trimpath -tags extended -o main .; \
    mkdir -p /tmp/site;\
    apk add --no-cache \
        binutils=2.39-r2 \
        grep=3.8-r1 \
        libgcc=12.2.1_git20220924-r4 \
        libstdc++=12.2.1_git20220924-r4 \
        musl=1.2.3-r4 \
        pcre2=10.42-r0 \
        zlib=1.2.13-r0 \
        ;

FROM reg.g5d.dev/baseglibc@sha256:5400aeaae743c557874709c0a0c5208e77d89582014989e51a13e7cae6b429e4

ENV HUGO_CACHEDIR="/build/cache"
COPY --from=build /src/hugo/main /usr/bin/hugo
COPY --from=build --chown=65532:65532 /tmp/site /build
WORKDIR /build
CMD ["/usr/bin/hugo", "-s", "/build/site", "-d", "/build/public"]
