# Generated from https://github.com/csmith/dockerfiles/blob/master/node/Dockerfile.gotpl
# BOM: {"apk:binutils":"2.35.2-r2","apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.33.1-r3","apk:ca-certificates":"20191127-r5","apk:expat":"2.4.1-r0","apk:g++":"10.3.1_git20210424-r2","apk:gcc":"10.3.1_git20210424-r2","apk:gdbm":"1.19-r0","apk:git":"2.32.0-r0","apk:gmp":"6.2.1-r0","apk:isl22":"0.22-r0","apk:libatomic":"10.3.1_git20210424-r2","apk:libbz2":"1.0.8-r1","apk:libc-dev":"0.7.2-r3","apk:libcrypto1.1":"1.1.1l-r0","apk:libcurl":"7.79.1-r0","apk:libffi":"3.3-r2","apk:libgcc":"10.3.1_git20210424-r2","apk:libgomp":"10.3.1_git20210424-r2","apk:libgphobos":"10.3.1_git20210424-r2","apk:libssl1.1":"1.1.1l-r0","apk:libstdc++":"10.3.1_git20210424-r2","apk:linux-headers":"5.10.41-r0","apk:make":"4.3-r0","apk:mpc1":"1.2.1-r0","apk:mpdecimal":"2.5.1-r1","apk:mpfr4":"4.1.0-r0","apk:musl":"1.2.2-r3","apk:musl-dev":"1.2.2-r3","apk:ncurses-libs":"6.2_p20210612-r0","apk:ncurses-terminfo-base":"6.2_p20210612-r0","apk:nghttp2-libs":"1.43.0-r0","apk:pcre2":"10.36-r0","apk:python3":"3.9.5-r1","apk:readline":"8.1.0-r0","apk:sqlite-libs":"3.35.5-r0","apk:xz-libs":"5.2.5-r0","apk:zlib":"1.2.11-r3","github:nodejs/node":"v16.10.0","image:alpine":"sha256:7cdc7a220adf9635167126cfe627b6ff000c12ddc5a44c7c5796161a81dae302","image:base":"sha256:1"}

FROM reg.g5d.dev/alpine@sha256:7cdc7a220adf9635167126cfe627b6ff000c12ddc5a44c7c5796161a81dae302 as build

ARG TAG="v16.10.0"

RUN set -eux; \
    apk add --no-cache \
        binutils=2.35.2-r2\
        brotli-libs=1.0.9-r5\
        busybox=1.33.1-r3\
        ca-certificates=20191127-r5\
        expat=2.4.1-r0\
        g++=10.3.1_git20210424-r2\
        gcc=10.3.1_git20210424-r2\
        gdbm=1.19-r0\
        git=2.32.0-r0\
        gmp=6.2.1-r0\
        isl22=0.22-r0\
        libatomic=10.3.1_git20210424-r2\
        libbz2=1.0.8-r1\
        libc-dev=0.7.2-r3\
        libcrypto1.1=1.1.1l-r0\
        libcurl=7.79.1-r0\
        libffi=3.3-r2\
        libgcc=10.3.1_git20210424-r2\
        libgomp=10.3.1_git20210424-r2\
        libgphobos=10.3.1_git20210424-r2\
        libssl1.1=1.1.1l-r0\
        libstdc++=10.3.1_git20210424-r2\
        linux-headers=5.10.41-r0\
        make=4.3-r0\
        mpc1=1.2.1-r0\
        mpdecimal=2.5.1-r1\
        mpfr4=4.1.0-r0\
        musl=1.2.2-r3\
        musl-dev=1.2.2-r3\
        ncurses-libs=6.2_p20210612-r0\
        ncurses-terminfo-base=6.2_p20210612-r0\
        nghttp2-libs=1.43.0-r0\
        pcre2=10.36-r0\
        python3=3.9.5-r1\
        readline=8.1.0-r0\
        sqlite-libs=3.35.5-r0\
        xz-libs=5.2.5-r0\
        zlib=1.2.11-r3\
        ;

RUN set -eux; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/nodejs/node /src; \
    cd /src; \
    ./configure --fully-static --enable-static; \
    make -j $(nproc); \
    make install

FROM reg.g5d.dev/base@sha256:5fc82ed384dc72f0fb8b4c2d57fc45ecdca45f1b4de6fee243464f030864854f

COPY --from=build /usr/local /usr/local

CMD ["/usr/local/node"]
