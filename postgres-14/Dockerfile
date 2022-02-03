# Generated from https://github.com/greboid/dockerfiles/blob/master/postgres-14/Dockerfile.gotpl
# BOM: {"apk:binutils":"2.37-r3","apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.34.1-r4","apk:ca-certificates":"20211220-r0","apk:curl":"7.80.0-r0","apk:g++":"10.3.1_git20211027-r0","apk:gcc":"10.3.1_git20211027-r0","apk:gmp":"6.2.1-r1","apk:isl22":"0.22-r0","apk:libatomic":"10.3.1_git20211027-r0","apk:libc-dev":"0.7.2-r3","apk:libcrypto1.1":"1.1.1l-r8","apk:libcurl":"7.80.0-r0","apk:libgcc":"10.3.1_git20211027-r0","apk:libgomp":"10.3.1_git20211027-r0","apk:libgphobos":"10.3.1_git20211027-r0","apk:libhistory":"8.1.1-r0","apk:libssl1.1":"1.1.1l-r8","apk:libstdc++":"10.3.1_git20211027-r0","apk:linux-headers":"5.10.41-r0","apk:make":"4.3-r0","apk:mpc1":"1.2.1-r0","apk:mpfr4":"4.1.0-r0","apk:musl":"1.2.2-r7","apk:musl-dev":"1.2.2-r7","apk:musl-locales":"0_git20201005-r0","apk:ncurses-dev":"6.3_p20211120-r0","apk:ncurses-libs":"6.3_p20211120-r0","apk:ncurses-terminfo-base":"6.3_p20211120-r0","apk:nghttp2-libs":"1.46.0-r0","apk:openssl-dev":"1.1.1l-r8","apk:pkgconf":"1.8.0-r0","apk:readline":"8.1.1-r0","apk:readline-dev":"8.1.1-r0","apk:zlib":"1.2.11-r3","apk:zlib-dev":"1.2.11-r3","image:alpine":"ef95559f475cdf56e55db71c168f1616d497e2215bbd9f6bc63398036ae546a1","image:base":"e390636a8773e2b84066cdbde990e80b3c97b8e7e25428df9f4b8f097323f861","postgres14":"14.1"}

FROM reg.g5d.dev/alpine@sha256:ef95559f475cdf56e55db71c168f1616d497e2215bbd9f6bc63398036ae546a1 as build

ARG ARCHIVE_URL="https://ftp.postgresql.org/pub/source/v14.1/postgresql-14.1.tar.bz2"
ARG ARCHIVE_SUM="4d3c101ea7ae38982f06bdc73758b53727fb6402ecd9382006fa5ecc7c2ca41f"

RUN set -eux; \
    apk add --no-cache \
        binutils=2.37-r3 \
        brotli-libs=1.0.9-r5 \
        busybox=1.34.1-r4 \
        ca-certificates=20211220-r0 \
        curl=7.80.0-r0 \
        g++=10.3.1_git20211027-r0 \
        gcc=10.3.1_git20211027-r0 \
        gmp=6.2.1-r1 \
        isl22=0.22-r0 \
        libatomic=10.3.1_git20211027-r0 \
        libc-dev=0.7.2-r3 \
        libcrypto1.1=1.1.1l-r8 \
        libcurl=7.80.0-r0 \
        libgcc=10.3.1_git20211027-r0 \
        libgomp=10.3.1_git20211027-r0 \
        libgphobos=10.3.1_git20211027-r0 \
        libhistory=8.1.1-r0 \
        libssl1.1=1.1.1l-r8 \
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
        openssl-dev=1.1.1l-r8 \
        pkgconf=1.8.0-r0 \
        readline=8.1.1-r0 \
        readline-dev=8.1.1-r0 \
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

FROM reg.g5d.dev/base@sha256:e390636a8773e2b84066cdbde990e80b3c97b8e7e25428df9f4b8f097323f861

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