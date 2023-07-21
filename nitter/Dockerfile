# Generated from https://github.com/greboid/dockerfiles/blob/master/nitter/Containerfile.gotpl
# BOM: {"apk:bash":"5.2.15-r5","apk:binutils":"2.40-r7","apk:brotli-libs":"1.0.9-r14","apk:busybox":"1.36.1-r1","apk:busybox-binsh":"1.36.1-r1","apk:ca-certificates":"20230506-r0","apk:curl":"8.2.0-r1","apk:gcc":"12.2.1_git20220924-r10","apk:git":"2.40.1-r0","apk:gmp":"6.2.1-r3","apk:isl25":"0.25-r2","apk:libacl":"2.3.1-r3","apk:libatomic":"12.2.1_git20220924-r10","apk:libcrypto3":"3.1.1-r3","apk:libcurl":"8.2.0-r1","apk:libexpat":"2.5.0-r1","apk:libgcc":"12.2.1_git20220924-r10","apk:libgomp":"12.2.1_git20220924-r10","apk:libidn2":"2.3.4-r1","apk:libncursesw":"6.4_p20230506-r0","apk:libpcre16":"8.45-r3","apk:libpcre32":"8.45-r3","apk:libpcrecpp":"8.45-r3","apk:libsass":"3.6.5-r0","apk:libsass-dev":"3.6.5-r0","apk:libssl3":"3.1.1-r3","apk:libstdc++":"12.2.1_git20220924-r10","apk:libunistring":"1.1-r1","apk:libxxhash":"0.8.1-r4","apk:lz4-libs":"1.9.4-r4","apk:mpc1":"1.3.1-r1","apk:mpfr4":"4.2.0_p12-r0","apk:musl":"1.2.4-r0","apk:musl-dev":"1.2.4-r0","apk:ncurses-terminfo-base":"6.4_p20230506-r0","apk:nghttp2-libs":"1.55.1-r0","apk:openssl":"3.1.1-r3","apk:openssl-dev":"3.1.1-r3","apk:pcre":"8.45-r3","apk:pcre-dev":"8.45-r3","apk:pcre2":"10.42-r1","apk:pkgconf":"1.9.5-r0","apk:popt":"1.19-r2","apk:readline":"8.2.1-r1","apk:rsync":"3.2.7-r4","apk:sqlite-libs":"3.41.2-r2","apk:zlib":"1.2.13-r1","apk:zstd-libs":"1.5.5-r4","github:nim-lang/nim":"v1.6.14","image:alpine":"347bebd1a359b0cccf62ebd4f65854ddf25331f9458011a2cc06a5e1679e3578","image:base":"5c298191cd12c92b8b3df64ab907f33118d5ddffa99f0c31d43c9623330c46e2"}

FROM reg.g5d.dev/alpine@sha256:347bebd1a359b0cccf62ebd4f65854ddf25331f9458011a2cc06a5e1679e3578 as nimbuild

ENV TAG=v1.6.14

ENV PATH=/opt/nim/bin:/opt/nimble/bin:$PATH

RUN set -eux; \
    # Dependencies and source for bootstrapping
    apk add --no-cache --virtual .build-deps \
        brotli-libs=1.0.9-r14 \
        busybox=1.36.1-r1 \
        busybox-binsh=1.36.1-r1 \
        ca-certificates=20230506-r0 \
        curl=8.2.0-r1 \
        libcrypto3=3.1.1-r3 \
        libcurl=8.2.0-r1 \
        libidn2=2.3.4-r1 \
        libssl3=3.1.1-r3 \
        libunistring=1.1-r1 \
        musl=1.2.4-r0 \
        nghttp2-libs=1.55.1-r0 \
        zlib=1.2.13-r1 \
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
        binutils=2.40-r7 \
        brotli-libs=1.0.9-r14 \
        busybox=1.36.1-r1 \
        busybox-binsh=1.36.1-r1 \
        ca-certificates=20230506-r0 \
        gcc=12.2.1_git20220924-r10 \
        git=2.40.1-r0 \
        gmp=6.2.1-r3 \
        isl25=0.25-r2 \
        libatomic=12.2.1_git20220924-r10 \
        libcrypto3=3.1.1-r3 \
        libcurl=8.2.0-r1 \
        libexpat=2.5.0-r1 \
        libgcc=12.2.1_git20220924-r10 \
        libgomp=12.2.1_git20220924-r10 \
        libidn2=2.3.4-r1 \
        libssl3=3.1.1-r3 \
        libstdc++=12.2.1_git20220924-r10 \
        libunistring=1.1-r1 \
        mpc1=1.3.1-r1 \
        mpfr4=4.2.0_p12-r0 \
        musl=1.2.4-r0 \
        musl-dev=1.2.4-r0 \
        nghttp2-libs=1.55.1-r0 \
        pcre2=10.42-r1 \
        zlib=1.2.13-r1 \
        zstd-libs=1.5.5-r4 \
        ;

FROM nimbuild as build

#No tags on the project, will just have to update this as needed until some tags are made
ENV TAG=bc219aa792cc0e4117888b2036a969559f4f2689

RUN set -eux; \
    apk add --no-cache \
        bash=5.2.15-r5 \
        binutils=2.40-r7 \
        brotli-libs=1.0.9-r14 \
        busybox=1.36.1-r1 \
        busybox-binsh=1.36.1-r1 \
        ca-certificates=20230506-r0 \
        gcc=12.2.1_git20220924-r10 \
        git=2.40.1-r0 \
        gmp=6.2.1-r3 \
        isl25=0.25-r2 \
        libacl=2.3.1-r3 \
        libatomic=12.2.1_git20220924-r10 \
        libcrypto3=3.1.1-r3 \
        libcurl=8.2.0-r1 \
        libexpat=2.5.0-r1 \
        libgcc=12.2.1_git20220924-r10 \
        libgomp=12.2.1_git20220924-r10 \
        libidn2=2.3.4-r1 \
        libncursesw=6.4_p20230506-r0 \
        libpcre16=8.45-r3 \
        libpcre32=8.45-r3 \
        libpcrecpp=8.45-r3 \
        libsass=3.6.5-r0 \
        libsass-dev=3.6.5-r0 \
        libssl3=3.1.1-r3 \
        libstdc++=12.2.1_git20220924-r10 \
        libunistring=1.1-r1 \
        libxxhash=0.8.1-r4 \
        lz4-libs=1.9.4-r4 \
        mpc1=1.3.1-r1 \
        mpfr4=4.2.0_p12-r0 \
        musl=1.2.4-r0 \
        musl-dev=1.2.4-r0 \
        ncurses-terminfo-base=6.4_p20230506-r0 \
        nghttp2-libs=1.55.1-r0 \
        openssl=3.1.1-r3 \
        openssl-dev=3.1.1-r3 \
        pcre=8.45-r3 \
        pcre-dev=8.45-r3 \
        pcre2=10.42-r1 \
        pkgconf=1.9.5-r0 \
        popt=1.19-r2 \
        readline=8.2.1-r1 \
        rsync=3.2.7-r4 \
        sqlite-libs=3.41.2-r2 \
        zlib=1.2.13-r1 \
        zstd-libs=1.5.5-r4 \
        ; \
    git clone -b master --single-branch https://github.com/zedeus/nitter /src/nitter; \
    cd /src/nitter; \
    git checkout $TAG; \
    nimble build -y -d:release; \
    strip -s nitter; \
    nimble scss; \
    nimble md; \
    mkdir -p /rootfs/home/nonroot/public; \
    cp /src/nitter/nitter /rootfs/home/nonroot/nitter; \
    rsync -ap /src/nitter/public/ /rootfs/home/nonroot/public/

RUN set -eux; \
    DEPS=$(scanelf --needed --nobanner --format '%n#p' --recursive -l /src/nitter/nitter \
    | tr ',' '\n' \
    | sort -u \
    | awk 'system("[ -e /usr/local/lib/" $1 " ]") == 0 { next } { print "so:" $1 }');\
    for DEP in $DEPS; do \
      apk add "${DEP}"; \
      apk info -qL "${DEP}" | rsync -aq '--files-from=-' / /rootfs/; \
    done;

FROM reg.g5d.dev/base@sha256:5c298191cd12c92b8b3df64ab907f33118d5ddffa99f0c31d43c9623330c46e2

copy --from=build --chown=65532:65532 /rootfs/ /
WORKDIR /home/nonroot
ENTRYPOINT ["/home/nonroot/nitter"]
