# Generated from https://github.com/greboid/dockerfiles/blob/master/nitter/Dockerfile.gotpl
# BOM: {"apk:bash":"5.2.15-r0","apk:binutils":"2.39-r2","apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r2","apk:curl":"7.86.0-r1","apk:gcc":"12.2.1_git20220924-r4","apk:git":"2.38.2-r0","apk:gmp":"6.2.1-r2","apk:isl24":"0.24-r1","apk:libatomic":"12.2.1_git20220924-r4","apk:libcrypto3":"3.0.7-r0","apk:libcurl":"7.86.0-r1","apk:libexpat":"2.5.0-r0","apk:libgcc":"12.2.1_git20220924-r4","apk:libgomp":"12.2.1_git20220924-r4","apk:libpcre16":"8.45-r2","apk:libpcre32":"8.45-r2","apk:libpcrecpp":"8.45-r2","apk:libsass":"3.6.5-r0","apk:libsass-dev":"3.6.5-r0","apk:libssl3":"3.0.7-r0","apk:libstdc++":"12.2.1_git20220924-r4","apk:mpc1":"1.2.1-r1","apk:mpfr4":"4.1.0-r0","apk:musl":"1.2.3-r4","apk:musl-dev":"1.2.3-r4","apk:ncurses-libs":"6.3_p20221119-r0","apk:ncurses-terminfo-base":"6.3_p20221119-r0","apk:nghttp2-libs":"1.51.0-r0","apk:openssl":"3.0.7-r0","apk:openssl-dev":"3.0.7-r0","apk:pcre":"8.45-r2","apk:pcre-dev":"8.45-r2","apk:pcre2":"10.42-r0","apk:pkgconf":"1.9.3-r0","apk:readline":"8.2.0-r0","apk:sqlite-libs":"3.40.0-r0","apk:zlib":"1.2.13-r0","github:nim-lang/nim":"v1.6.10","image:alpine":"3a97af24094652229c76b9dd8dd2895cf50ea52a22b1725db085070041e016ce","image:base":"ffd04efedafc54f5392c35a1045604c12a4329e5780407d5fe08dfe495d411b1"}

FROM reg.g5d.dev/alpine@sha256:3a97af24094652229c76b9dd8dd2895cf50ea52a22b1725db085070041e016ce as nimbuild

ENV TAG=v1.6.10

ENV PATH=/opt/nim/bin:/opt/nimble/bin:$PATH

RUN set -eux; \
    # Dependencies and source for bootstrapping
    apk add --no-cache --virtual .build-deps \
        brotli-libs=1.0.9-r9 \
        busybox=1.35.0-r29 \
        busybox-binsh=1.35.0-r29 \
        ca-certificates=20220614-r2 \
        curl=7.86.0-r1 \
        libcrypto3=3.0.7-r0 \
        libcurl=7.86.0-r1 \
        libssl3=3.0.7-r0 \
        musl=1.2.3-r4 \
        nghttp2-libs=1.51.0-r0 \
        zlib=1.2.13-r0 \
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
        binutils=2.39-r2 \
        brotli-libs=1.0.9-r9 \
        busybox=1.35.0-r29 \
        busybox-binsh=1.35.0-r29 \
        ca-certificates=20220614-r2 \
        gcc=12.2.1_git20220924-r4 \
        git=2.38.2-r0 \
        gmp=6.2.1-r2 \
        isl24=0.24-r1 \
        libatomic=12.2.1_git20220924-r4 \
        libcrypto3=3.0.7-r0 \
        libcurl=7.86.0-r1 \
        libexpat=2.5.0-r0 \
        libgcc=12.2.1_git20220924-r4 \
        libgomp=12.2.1_git20220924-r4 \
        libssl3=3.0.7-r0 \
        libstdc++=12.2.1_git20220924-r4 \
        mpc1=1.2.1-r1 \
        mpfr4=4.1.0-r0 \
        musl=1.2.3-r4 \
        musl-dev=1.2.3-r4 \
        nghttp2-libs=1.51.0-r0 \
        pcre2=10.42-r0 \
        zlib=1.2.13-r0 \
        ;

FROM nimbuild as build

#No tags on the project, will just have to update this as needed until some tags are made
ENV TAG=2ac3afa5b273a502d7632e9346c7c3bc9283fb48

RUN set -eux; \
    apk add --no-cache \
        bash=5.2.15-r0 \
        binutils=2.39-r2 \
        brotli-libs=1.0.9-r9 \
        busybox=1.35.0-r29 \
        busybox-binsh=1.35.0-r29 \
        ca-certificates=20220614-r2 \
        gcc=12.2.1_git20220924-r4 \
        git=2.38.2-r0 \
        gmp=6.2.1-r2 \
        isl24=0.24-r1 \
        libatomic=12.2.1_git20220924-r4 \
        libcrypto3=3.0.7-r0 \
        libcurl=7.86.0-r1 \
        libexpat=2.5.0-r0 \
        libgcc=12.2.1_git20220924-r4 \
        libgomp=12.2.1_git20220924-r4 \
        libpcre16=8.45-r2 \
        libpcre32=8.45-r2 \
        libpcrecpp=8.45-r2 \
        libsass=3.6.5-r0 \
        libsass-dev=3.6.5-r0 \
        libssl3=3.0.7-r0 \
        libstdc++=12.2.1_git20220924-r4 \
        mpc1=1.2.1-r1 \
        mpfr4=4.1.0-r0 \
        musl=1.2.3-r4 \
        musl-dev=1.2.3-r4 \
        ncurses-libs=6.3_p20221119-r0 \
        ncurses-terminfo-base=6.3_p20221119-r0 \
        nghttp2-libs=1.51.0-r0 \
        openssl=3.0.7-r0 \
        openssl-dev=3.0.7-r0 \
        pcre=8.45-r2 \
        pcre-dev=8.45-r2 \
        pcre2=10.42-r0 \
        pkgconf=1.9.3-r0 \
        readline=8.2.0-r0 \
        sqlite-libs=3.40.0-r0 \
        zlib=1.2.13-r0 \
        ; \
    git clone -b master --single-branch https://github.com/zedeus/nitter /src/nitter; \
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

FROM reg.g5d.dev/base@sha256:ffd04efedafc54f5392c35a1045604c12a4329e5780407d5fe08dfe495d411b1

COPY --from=build /src/libs/* /usr/lib/
COPY --from=build /src/nitter/nitter /home/nonroot/nitter
COPY --from=build /src/nitter/public /home/nonroot/public

WORKDIR /home/nonroot
ENTRYPOINT ["/home/nonroot/nitter"]
