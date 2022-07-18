# Generated from https://github.com/greboid/dockerfiles/blob/master/postgres-14/Dockerfile.gotpl
# BOM: {"apk:binutils":"2.38-r3","apk:brotli-libs":"1.0.9-r6","apk:busybox":"1.35.0-r15","apk:ca-certificates":"20211220-r0","apk:curl":"7.83.1-r2","apk:g++":"11.2.1_git20220219-r2","apk:gcc":"11.2.1_git20220219-r2","apk:gmp":"6.2.1-r2","apk:isl22":"0.22-r0","apk:libatomic":"11.2.1_git20220219-r2","apk:libc-dev":"0.7.2-r3","apk:libcrypto1.1":"1.1.1q-r0","apk:libcurl":"7.83.1-r2","apk:libgcc":"11.2.1_git20220219-r2","apk:libgomp":"11.2.1_git20220219-r2","apk:libhistory":"8.1.2-r0","apk:libssl1.1":"1.1.1q-r0","apk:libstdc++":"11.2.1_git20220219-r2","apk:linux-headers":"5.16.7-r1","apk:make":"4.3-r0","apk:mpc1":"1.2.1-r0","apk:mpfr4":"4.1.0-r0","apk:musl":"1.2.3-r0","apk:musl-dev":"1.2.3-r0","apk:musl-locales":"0.1.0-r0","apk:ncurses-dev":"6.3_p20220521-r0","apk:ncurses-libs":"6.3_p20220521-r0","apk:ncurses-terminfo-base":"6.3_p20220521-r0","apk:nghttp2-libs":"1.47.0-r0","apk:openssl-dev":"1.1.1q-r0","apk:pkgconf":"1.8.0-r0","apk:readline":"8.1.2-r0","apk:readline-dev":"8.1.2-r0","apk:zlib":"1.2.12-r1","apk:zlib-dev":"1.2.12-r1","image:alpine":"e1af61331d46ce7a5770c5083e1c81e297fbf3fcc97b63542b0ac62a8f3900a3","image:base":"d8e173562b2af5c2d2a6bedc5af4a36a9fe9e9561aca25d22c947752a1695a64","postgres14":"14.4"}

FROM reg.g5d.dev/alpine@sha256:e1af61331d46ce7a5770c5083e1c81e297fbf3fcc97b63542b0ac62a8f3900a3 as build

ARG ARCHIVE_URL="https://ftp.postgresql.org/pub/source/v14.4/postgresql-14.4.tar.bz2"
ARG ARCHIVE_SUM="c23b6237c5231c791511bdc79098617d6852e9e3bdf360efd8b5d15a1a3d8f6a"

RUN set -eux; \
    apk add --no-cache \
        binutils=2.38-r3 \
        brotli-libs=1.0.9-r6 \
        busybox=1.35.0-r15 \
        ca-certificates=20211220-r0 \
        curl=7.83.1-r2 \
        g++=11.2.1_git20220219-r2 \
        gcc=11.2.1_git20220219-r2 \
        gmp=6.2.1-r2 \
        isl22=0.22-r0 \
        libatomic=11.2.1_git20220219-r2 \
        libc-dev=0.7.2-r3 \
        libcrypto1.1=1.1.1q-r0 \
        libcurl=7.83.1-r2 \
        libgcc=11.2.1_git20220219-r2 \
        libgomp=11.2.1_git20220219-r2 \
        libhistory=8.1.2-r0 \
        libssl1.1=1.1.1q-r0 \
        libstdc++=11.2.1_git20220219-r2 \
        linux-headers=5.16.7-r1 \
        make=4.3-r0 \
        mpc1=1.2.1-r0 \
        mpfr4=4.1.0-r0 \
        musl=1.2.3-r0 \
        musl-dev=1.2.3-r0 \
        musl-locales=0.1.0-r0 \
        ncurses-dev=6.3_p20220521-r0 \
        ncurses-libs=6.3_p20220521-r0 \
        ncurses-terminfo-base=6.3_p20220521-r0 \
        nghttp2-libs=1.47.0-r0 \
        openssl-dev=1.1.1q-r0 \
        pkgconf=1.8.0-r0 \
        readline=8.1.2-r0 \
        readline-dev=8.1.2-r0 \
        zlib=1.2.12-r1 \
        zlib-dev=1.2.12-r1 \
        ; \
    curl -sfLo postgresql.tar.bz2 $ARCHIVE_URL; \
    echo "$ARCHIVE_SUM *postgresql.tar.bz2" | sha256sum -wc -; \
    mkdir -p /usr/local/postgres; \
    tar -C /usr/local/postgres --strip-components 1 -xf postgresql.tar.bz2; \
    cd /usr/local/postgres; \
    ./configure \
       --with-system-tzdata=/usr/share/zoneinfo \
        --with-ssl=openssl; \
    make -j$(nproc); \
    make install; \
    mkdir -p /src/libs; \
    cp /usr/lib/libreadline* /src/libs/; \
    cp /usr/lib/libncursesw* /src/libs/; \
    cp /usr/lib/libcrypto* /src/libs/; \
    cp /usr/lib/libssl* /src/libs/; \
    mkdir /src/data

FROM reg.g5d.dev/base@sha256:d8e173562b2af5c2d2a6bedc5af4a36a9fe9e9561aca25d22c947752a1695a64

ENV LANG en_US.utf8
ENV PGDATA=/var/lib/postgres/data

COPY --from=build /bin/sh /bin/sh
COPY --from=build /usr/bin/locale /usr/bin/locale
COPY --from=build /src/libs /usr/lib/
COPY --from=build /usr/local/pgsql /usr/local/
COPY --from=build --chown=65532:65532 /src/data /var/lib/postgres
COPY postgresql.conf /etc/postgresql.conf
COPY entry.sh /usr/local/bin/

STOPSIGNAL SIGINT

ENTRYPOINT ["/usr/local/bin/entry.sh"]