# Generated from https://github.com/greboid/dockerfiles/blob/master/nitter/Dockerfile.gotpl
# BOM: {"apk:bash":"5.1.16-r2","apk:binutils":"2.38-r3","apk:brotli-libs":"1.0.9-r6","apk:busybox":"1.35.0-r15","apk:ca-certificates":"20211220-r0","apk:curl":"7.83.1-r2","apk:expat":"2.4.8-r0","apk:gcc":"11.2.1_git20220219-r2","apk:git":"2.36.2-r0","apk:gmp":"6.2.1-r2","apk:isl22":"0.22-r0","apk:libatomic":"11.2.1_git20220219-r2","apk:libcrypto1.1":"1.1.1q-r0","apk:libcurl":"7.83.1-r2","apk:libgcc":"11.2.1_git20220219-r2","apk:libgomp":"11.2.1_git20220219-r2","apk:libpcre16":"8.45-r2","apk:libpcre32":"8.45-r2","apk:libpcrecpp":"8.45-r2","apk:libsass":"3.6.5-r0","apk:libsass-dev":"3.6.5-r0","apk:libssl1.1":"1.1.1q-r0","apk:libstdc++":"11.2.1_git20220219-r2","apk:mpc1":"1.2.1-r0","apk:mpfr4":"4.1.0-r0","apk:musl":"1.2.3-r0","apk:musl-dev":"1.2.3-r0","apk:ncurses-libs":"6.3_p20220521-r0","apk:ncurses-terminfo-base":"6.3_p20220521-r0","apk:nghttp2-libs":"1.47.0-r0","apk:openssl":"1.1.1q-r0","apk:openssl-dev":"1.1.1q-r0","apk:pcre":"8.45-r2","apk:pcre-dev":"8.45-r2","apk:pcre2":"10.40-r0","apk:pkgconf":"1.8.0-r0","apk:readline":"8.1.2-r0","apk:sqlite-libs":"3.38.5-r0","apk:zlib":"1.2.12-r1","github:nim-lang/nim":"v1.6.6","image:alpine":"e1af61331d46ce7a5770c5083e1c81e297fbf3fcc97b63542b0ac62a8f3900a3","image:base":"d8e173562b2af5c2d2a6bedc5af4a36a9fe9e9561aca25d22c947752a1695a64"}

FROM reg.g5d.dev/alpine@sha256:e1af61331d46ce7a5770c5083e1c81e297fbf3fcc97b63542b0ac62a8f3900a3 as nimbuild

ENV TAG=v1.6.6

ENV PATH=/opt/nim/bin:/opt/nimble/bin:$PATH

RUN set -eux; \
    # Dependencies and source for bootstrapping
    apk add --no-cache --virtual .build-deps \
        brotli-libs=1.0.9-r6 \
        busybox=1.35.0-r15 \
        ca-certificates=20211220-r0 \
        curl=7.83.1-r2 \
        libcrypto1.1=1.1.1q-r0 \
        libcurl=7.83.1-r2 \
        libssl1.1=1.1.1q-r0 \
        musl=1.2.3-r0 \
        nghttp2-libs=1.47.0-r0 \
        zlib=1.2.12-r1 \
        ; \
    curl -sfLo nim.tar.xz https://nim-lang.org/download/nim-${TAG:1}-linux_x64.tar.xz; \
    curl -sfLo nim.tar.xz.sha256 https://nim-lang.org/download/nim-${TAG:1}-linux_x64.tar.xz.sha256; \
    echo "$(awk '{print $1}' nim.tar.xz.sha256) *nim.tar.xz" | sha256sum -wc -; \
    tar -C /tmp -xf nim.tar.xz; \
    mv /tmp/nim-* /tmp/nim; \
    cd /tmp/nim; \
    ./install.sh /usr/bin; \
    cp /tmp/nim/bin/nimble /usr/bin/nimble; \
    apk del --no-network .build-deps; \
    rm -rf \
        /nim.tar.* \
        /tmp/nim; \
    apk add --no-cache \
        binutils=2.38-r3 \
        brotli-libs=1.0.9-r6 \
        busybox=1.35.0-r15 \
        ca-certificates=20211220-r0 \
        expat=2.4.8-r0 \
        gcc=11.2.1_git20220219-r2 \
        git=2.36.2-r0 \
        gmp=6.2.1-r2 \
        isl22=0.22-r0 \
        libatomic=11.2.1_git20220219-r2 \
        libcrypto1.1=1.1.1q-r0 \
        libcurl=7.83.1-r2 \
        libgcc=11.2.1_git20220219-r2 \
        libgomp=11.2.1_git20220219-r2 \
        libssl1.1=1.1.1q-r0 \
        libstdc++=11.2.1_git20220219-r2 \
        mpc1=1.2.1-r0 \
        mpfr4=4.1.0-r0 \
        musl=1.2.3-r0 \
        musl-dev=1.2.3-r0 \
        nghttp2-libs=1.47.0-r0 \
        pcre2=10.40-r0 \
        zlib=1.2.12-r1 \
        ;

FROM nimbuild as build

#No tags on the project, will just have to update this as needed until some tags are made
ENV TAG=6884f05041a9b8619ec709afacdfdd6482a120a0

RUN set -eux; \
    apk add --no-cache \
        bash=5.1.16-r2 \
        binutils=2.38-r3 \
        brotli-libs=1.0.9-r6 \
        busybox=1.35.0-r15 \
        ca-certificates=20211220-r0 \
        expat=2.4.8-r0 \
        gcc=11.2.1_git20220219-r2 \
        git=2.36.2-r0 \
        gmp=6.2.1-r2 \
        isl22=0.22-r0 \
        libatomic=11.2.1_git20220219-r2 \
        libcrypto1.1=1.1.1q-r0 \
        libcurl=7.83.1-r2 \
        libgcc=11.2.1_git20220219-r2 \
        libgomp=11.2.1_git20220219-r2 \
        libpcre16=8.45-r2 \
        libpcre32=8.45-r2 \
        libpcrecpp=8.45-r2 \
        libsass=3.6.5-r0 \
        libsass-dev=3.6.5-r0 \
        libssl1.1=1.1.1q-r0 \
        libstdc++=11.2.1_git20220219-r2 \
        mpc1=1.2.1-r0 \
        mpfr4=4.1.0-r0 \
        musl=1.2.3-r0 \
        musl-dev=1.2.3-r0 \
        ncurses-libs=6.3_p20220521-r0 \
        ncurses-terminfo-base=6.3_p20220521-r0 \
        nghttp2-libs=1.47.0-r0 \
        openssl=1.1.1q-r0 \
        openssl-dev=1.1.1q-r0 \
        pcre=8.45-r2 \
        pcre-dev=8.45-r2 \
        pcre2=10.40-r0 \
        pkgconf=1.8.0-r0 \
        readline=8.1.2-r0 \
        sqlite-libs=3.38.5-r0 \
        zlib=1.2.12-r1 \
        ; \
    git -c advice.detachedHead=false clone -b master --single-branch https://github.com/zedeus/nitter /src/nitter; \
    mkdir -p /src/libs; \
    cp /usr/lib/libsqlite3* /src/libs; \
    cp /usr/lib/libssl* /src/libs; \
    cp /usr/lib/libcrypto* /src/libs; \
    cp /usr/lib/libpcre* /src/libs; \
    cd /src/nitter; \
    git checkout $TAG; \
    nimble build -y -d:release; \
    strip -s nitter; \
    nimble scss

FROM reg.g5d.dev/base@sha256:d8e173562b2af5c2d2a6bedc5af4a36a9fe9e9561aca25d22c947752a1695a64

COPY --from=build /src/libs/* /usr/lib/
COPY --from=build /src/nitter/nitter /home/nonroot/nitter
COPY --from=build /src/nitter/public /home/nonroot/public

WORKDIR /home/nonroot
ENTRYPOINT ["/home/nonroot/nitter"]
