# Generated from https://github.com/greboid/dockerfiles/blob/master/postgres-14/Dockerfile.gotpl
# BOM: {"apk:binutils":"2.37-r3","apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.34.1-r5","apk:ca-certificates":"20211220-r0","apk:curl":"7.80.0-r1","apk:g++":"10.3.1_git20211027-r0","apk:gcc":"10.3.1_git20211027-r0","apk:gmp":"6.2.1-r1","apk:isl22":"0.22-r0","apk:libatomic":"10.3.1_git20211027-r0","apk:libc-dev":"0.7.2-r3","apk:libcrypto1.1":"1.1.1o-r0","apk:libcurl":"7.80.0-r1","apk:libgcc":"10.3.1_git20211027-r0","apk:libgomp":"10.3.1_git20211027-r0","apk:libgphobos":"10.3.1_git20211027-r0","apk:libhistory":"8.1.1-r0","apk:libssl1.1":"1.1.1o-r0","apk:libstdc++":"10.3.1_git20211027-r0","apk:linux-headers":"5.10.41-r0","apk:make":"4.3-r0","apk:mpc1":"1.2.1-r0","apk:mpfr4":"4.1.0-r0","apk:musl":"1.2.2-r7","apk:musl-dev":"1.2.2-r7","apk:musl-locales":"0_git20201005-r0","apk:ncurses-dev":"6.3_p20211120-r0","apk:ncurses-libs":"6.3_p20211120-r0","apk:ncurses-terminfo-base":"6.3_p20211120-r0","apk:nghttp2-libs":"1.46.0-r0","apk:openssl-dev":"1.1.1o-r0","apk:pkgconf":"1.8.0-r0","apk:readline":"8.1.1-r0","apk:readline-dev":"8.1.1-r0","apk:zlib":"1.2.12-r1","apk:zlib-dev":"1.2.12-r1","image:alpine":"9f06ed7307f2d03a3d486b9b0bfe7737dfe3c7b3e2f8e906f5a2fcc9408867ac","image:base":"ee8879a758bbe46a6b0b5b9d51a3cd3cb5660106d185b0f1fd981e4475663d08","postgres14":"14.3"}

FROM reg.g5d.dev/alpine@sha256:9f06ed7307f2d03a3d486b9b0bfe7737dfe3c7b3e2f8e906f5a2fcc9408867ac as build

ARG ARCHIVE_URL="https://ftp.postgresql.org/pub/source/v14.3/postgresql-14.3.tar.bz2"
ARG ARCHIVE_SUM="279057368bf59a919c05ada8f95c5e04abb43e74b9a2a69c3d46a20e07a9af38"

RUN set -eux; \
    apk add --no-cache \
        binutils=2.37-r3 \
        brotli-libs=1.0.9-r5 \
        busybox=1.34.1-r5 \
        ca-certificates=20211220-r0 \
        curl=7.80.0-r1 \
        g++=10.3.1_git20211027-r0 \
        gcc=10.3.1_git20211027-r0 \
        gmp=6.2.1-r1 \
        isl22=0.22-r0 \
        libatomic=10.3.1_git20211027-r0 \
        libc-dev=0.7.2-r3 \
        libcrypto1.1=1.1.1o-r0 \
        libcurl=7.80.0-r1 \
        libgcc=10.3.1_git20211027-r0 \
        libgomp=10.3.1_git20211027-r0 \
        libgphobos=10.3.1_git20211027-r0 \
        libhistory=8.1.1-r0 \
        libssl1.1=1.1.1o-r0 \
        libstdc++=10.3.1_git20211027-r0 \
        linux-headers=5.10.41-r0 \
        make=4.3-r0 \
        mpc1=1.2.1-r0 \
        mpfr4=4.1.0-r0 \
        musl=1.2.2-r7 \
        musl-dev=1.2.2-r7 \
        musl-locales=0_git20201005-r0 \
        ncurses-dev=6.3_p20211120-r0 \
        ncurses-libs=6.3_p20211120-r0 \
        ncurses-terminfo-base=6.3_p20211120-r0 \
        nghttp2-libs=1.46.0-r0 \
        openssl-dev=1.1.1o-r0 \
        pkgconf=1.8.0-r0 \
        readline=8.1.1-r0 \
        readline-dev=8.1.1-r0 \
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

FROM reg.g5d.dev/base@sha256:ee8879a758bbe46a6b0b5b9d51a3cd3cb5660106d185b0f1fd981e4475663d08

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