# Generated from https://github.com/greboid/dockerfiles/blob/master/postgres-14/Dockerfile.gotpl
# BOM: {"apk:binutils":"2.35.2-r2","apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.33.1-r3","apk:ca-certificates":"20191127-r5","apk:curl":"7.79.1-r0","apk:g++":"10.3.1_git20210424-r2","apk:gcc":"10.3.1_git20210424-r2","apk:gmp":"6.2.1-r0","apk:isl22":"0.22-r0","apk:libatomic":"10.3.1_git20210424-r2","apk:libc-dev":"0.7.2-r3","apk:libcrypto1.1":"1.1.1l-r0","apk:libcurl":"7.79.1-r0","apk:libgcc":"10.3.1_git20210424-r2","apk:libgomp":"10.3.1_git20210424-r2","apk:libgphobos":"10.3.1_git20210424-r2","apk:libhistory":"8.1.0-r0","apk:libssl1.1":"1.1.1l-r0","apk:libstdc++":"10.3.1_git20210424-r2","apk:linux-headers":"5.10.41-r0","apk:make":"4.3-r0","apk:mpc1":"1.2.1-r0","apk:mpfr4":"4.1.0-r0","apk:musl":"1.2.2-r3","apk:musl-dev":"1.2.2-r3","apk:musl-locales":"0_git20201005-r0","apk:ncurses-dev":"6.2_p20210612-r0","apk:ncurses-libs":"6.2_p20210612-r0","apk:ncurses-terminfo-base":"6.2_p20210612-r0","apk:nghttp2-libs":"1.43.0-r0","apk:openssl-dev":"1.1.1l-r0","apk:pkgconf":"1.7.4-r0","apk:readline":"8.1.0-r0","apk:readline-dev":"8.1.0-r0","apk:zlib":"1.2.11-r3","apk:zlib-dev":"1.2.11-r3","image:alpine":"sha256:c8797cc5b046929decf47059cc3a62b89d664ed442288d10833f496d1339af7f","image:base":"sha256:82873fbcddc94e3cf77fdfe36765391b6e6049701623a62c2a23248d2a42b1cf","postgres14":"14.0"}

FROM reg.g5d.dev/alpine@sha256:c8797cc5b046929decf47059cc3a62b89d664ed442288d10833f496d1339af7f as build

ARG ARCHIVE_URL="https://ftp.postgresql.org/pub/source/v14.0/postgresql-14.0.tar.bz2"
ARG ARCHIVE_SUM="ee2ad79126a7375e9102c4db77c4acae6ae6ffe3e082403b88826d96d927a122"

RUN set -eux; \
    apk add --no-cache \
        binutils=2.35.2-r2 \
        brotli-libs=1.0.9-r5 \
        busybox=1.33.1-r3 \
        ca-certificates=20191127-r5 \
        curl=7.79.1-r0 \
        g++=10.3.1_git20210424-r2 \
        gcc=10.3.1_git20210424-r2 \
        gmp=6.2.1-r0 \
        isl22=0.22-r0 \
        libatomic=10.3.1_git20210424-r2 \
        libc-dev=0.7.2-r3 \
        libcrypto1.1=1.1.1l-r0 \
        libcurl=7.79.1-r0 \
        libgcc=10.3.1_git20210424-r2 \
        libgomp=10.3.1_git20210424-r2 \
        libgphobos=10.3.1_git20210424-r2 \
        libhistory=8.1.0-r0 \
        libssl1.1=1.1.1l-r0 \
        libstdc++=10.3.1_git20210424-r2 \
        linux-headers=5.10.41-r0 \
        make=4.3-r0 \
        mpc1=1.2.1-r0 \
        mpfr4=4.1.0-r0 \
        musl=1.2.2-r3 \
        musl-dev=1.2.2-r3 \
        musl-locales=0_git20201005-r0 \
        ncurses-dev=6.2_p20210612-r0 \
        ncurses-libs=6.2_p20210612-r0 \
        ncurses-terminfo-base=6.2_p20210612-r0 \
        nghttp2-libs=1.43.0-r0 \
        openssl-dev=1.1.1l-r0 \
        pkgconf=1.7.4-r0 \
        readline=8.1.0-r0 \
        readline-dev=8.1.0-r0 \
        zlib=1.2.11-r3 \
        zlib-dev=1.2.11-r3 \
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

FROM reg.g5d.dev/base@sha256:82873fbcddc94e3cf77fdfe36765391b6e6049701623a62c2a23248d2a42b1cf

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