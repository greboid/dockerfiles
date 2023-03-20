# Generated from https://github.com/greboid/dockerfiles/blob/master/postgres-15/Dockerfile.gotpl
# BOM: {"apk:bash":"5.2.15-r0","apk:binutils":"2.39-r2","apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r4","apk:curl":"7.88.1-r1","apk:g++":"12.2.1_git20220924-r4","apk:gcc":"12.2.1_git20220924-r4","apk:git":"2.38.4-r1","apk:gmp":"6.2.1-r2","apk:isl24":"0.24-r1","apk:libacl":"2.3.1-r1","apk:libatomic":"12.2.1_git20220924-r4","apk:libc-dev":"0.7.2-r3","apk:libcrypto3":"3.0.8-r0","apk:libcurl":"7.88.1-r1","apk:libexpat":"2.5.0-r0","apk:libgcc":"12.2.1_git20220924-r4","apk:libgomp":"12.2.1_git20220924-r4","apk:libhistory":"8.2.0-r0","apk:libssl3":"3.0.8-r0","apk:libstdc++":"12.2.1_git20220924-r4","apk:libstdc++-dev":"12.2.1_git20220924-r4","apk:linux-headers":"5.19.5-r0","apk:lz4-libs":"1.9.4-r1","apk:make":"4.3-r1","apk:mpc1":"1.2.1-r1","apk:mpfr4":"4.1.0-r0","apk:musl":"1.2.3-r4","apk:musl-dev":"1.2.3-r4","apk:musl-locales":"0.1.0-r0","apk:ncurses-dev":"6.3_p20221119-r0","apk:ncurses-libs":"6.3_p20221119-r0","apk:ncurses-terminfo-base":"6.3_p20221119-r0","apk:nghttp2-libs":"1.51.0-r0","apk:openssl-dev":"3.0.8-r0","apk:pcre2":"10.42-r0","apk:pkgconf":"1.9.4-r0","apk:popt":"1.19-r0","apk:readline":"8.2.0-r0","apk:readline-dev":"8.2.0-r0","apk:rsync":"3.2.7-r0","apk:scanelf":"1.3.5-r1","apk:zlib":"1.2.13-r0","apk:zlib-dev":"1.2.13-r0","apk:zstd-libs":"1.5.2-r9","image:alpine":"e3385453363598236b7e826530c55b70d4a1a3e53931d4e7b675d47c4c5c8f8d","image:baseroot":"fd51da47fdc6d0f778cc71f985d36559355d00bfbefdb6b92718ca47a594760f","postgres15":"15.2"}

FROM reg.g5d.dev/alpine@sha256:e3385453363598236b7e826530c55b70d4a1a3e53931d4e7b675d47c4c5c8f8d as build

ARG ARCHIVE_URL="https://ftp.postgresql.org/pub/source/v15.2/postgresql-15.2.tar.bz2"
ARG ARCHIVE_SUM="99a2171fc3d6b5b5f56b757a7a3cb85d509a38e4273805def23941ed2b8468c7"
ARG ENTRYPOINT=39c3dc7046f5f7909179a51e25f2a61158781d57

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r9 \
        busybox=1.35.0-r29 \
        busybox-binsh=1.35.0-r29 \
        ca-certificates=20220614-r4 \
        git=2.38.4-r1 \
        libcrypto3=3.0.8-r0 \
        libcurl=7.88.1-r1 \
        libexpat=2.5.0-r0 \
        libssl3=3.0.8-r0 \
        musl=1.2.3-r4 \
        nghttp2-libs=1.51.0-r0 \
        pcre2=10.42-r0 \
        zlib=1.2.13-r0 \
        ; \
    git clone -b master --single-branch https://github.com/docker-library/postgres /src/docker-postgres; \
    cd /src/docker-postgres; \
    git checkout $ENTRYPOINT;

RUN set -eux; \
    apk add --no-cache \
        bash=5.2.15-r0 \
        binutils=2.39-r2 \
        brotli-libs=1.0.9-r9 \
        busybox=1.35.0-r29 \
        busybox-binsh=1.35.0-r29 \
        ca-certificates=20220614-r4 \
        curl=7.88.1-r1 \
        g++=12.2.1_git20220924-r4 \
        gcc=12.2.1_git20220924-r4 \
        gmp=6.2.1-r2 \
        isl24=0.24-r1 \
        libacl=2.3.1-r1 \
        libatomic=12.2.1_git20220924-r4 \
        libc-dev=0.7.2-r3 \
        libcrypto3=3.0.8-r0 \
        libcurl=7.88.1-r1 \
        libgcc=12.2.1_git20220924-r4 \
        libgomp=12.2.1_git20220924-r4 \
        libhistory=8.2.0-r0 \
        libssl3=3.0.8-r0 \
        libstdc++=12.2.1_git20220924-r4 \
        libstdc++-dev=12.2.1_git20220924-r4 \
        linux-headers=5.19.5-r0 \
        lz4-libs=1.9.4-r1 \
        make=4.3-r1 \
        mpc1=1.2.1-r1 \
        mpfr4=4.1.0-r0 \
        musl=1.2.3-r4 \
        musl-dev=1.2.3-r4 \
        musl-locales=0.1.0-r0 \
        ncurses-dev=6.3_p20221119-r0 \
        ncurses-libs=6.3_p20221119-r0 \
        ncurses-terminfo-base=6.3_p20221119-r0 \
        nghttp2-libs=1.51.0-r0 \
        openssl-dev=3.0.8-r0 \
        pkgconf=1.9.4-r0 \
        popt=1.19-r0 \
        readline=8.2.0-r0 \
        readline-dev=8.2.0-r0 \
        rsync=3.2.7-r0 \
        scanelf=1.3.5-r1 \
        zlib=1.2.13-r0 \
        zlib-dev=1.2.13-r0 \
        zstd-libs=1.5.2-r9 \
        ; \
    curl -sfLo postgresql.tar.bz2 $ARCHIVE_URL; \
    echo "$ARCHIVE_SUM *postgresql.tar.bz2" | sha256sum -wc -; \
    mkdir -p /usr/local/postgres; \
    tar -C /usr/local/postgres --strip-components 1 -xf postgresql.tar.bz2; \
    cd /usr/local/postgres; \
    ./configure \
        --with-system-tzdata=/usr/share/zoneinfo \
        --with-includes=/usr/local/include \
        --with-libraries=/usr/local/lib \
        --with-ssl=openssl; \
    make -j$(nproc) world; \
    make install-world; \
    make -C contrib install; \
    mkdir -p /src/data; \
    mkdir -p /src/run; \
    chmod 777 /src/run; \
    mkdir -p /etc/postgresql/; \
    mkdir -p /docker-entrypoint-initdb.d/;

RUN set -eux; \
    DEPS=$(scanelf --needed --nobanner --format '%n#p' --recursive -l /usr/local /usr/local/bin/psql /usr/local/bin/postgres /usr/local/bin/initdb \
    | tr ',' '\n' \
    | sort -u \
    | awk 'system("[ -e /usr/local/lib/" $1 " ]") == 0 { next } { print "so:" $1 }');\
    mkdir /rootfs; \
    for DEP in $DEPS; do \
      apk add "${DEP}"; \
      apk info -qL "${DEP}" | rsync -aq '--files-from=-' / /rootfs/; \
    done;

RUN ls -lR /rootfs/

FROM reg.g5d.dev/baseroot@sha256:fd51da47fdc6d0f778cc71f985d36559355d00bfbefdb6b92718ca47a594760f

ENV LANG en_US.utf8
ENV PGDATA=/var/lib/postgresql/data
ENV PGHOST=127.0.0.1
ENV PGUSER=postgres

COPY --from=build /rootfs/ /
COPY --from=build /usr/bin/env /usr/bin/env
COPY --from=build /bin/sh /bin/sh
COPY --from=build /bin/bash /bin/bash
COPY --from=build /usr/bin/locale /usr/bin/locale
COPY --from=build /bin/ls /bin/ls
COPY --from=build /usr/bin/id /usr/bin/id
COPY --from=build /bin/mkdir /bin/mkdir
COPY --from=build /bin/chmod /bin/chmod
COPY --from=build /usr/local/pgsql /usr/local/
COPY --from=build --chown=65532:65532 --chmod=750 /src/data /var/lib/postgresql
COPY --from=build --chown=65532:65532 --chmod=750 /src/data /var/lib/postgresql/data
COPY --from=build --chown=65532:65532 --chmod=777 /src/run /var/run/psql
COPY --from=build --chown=65532:65532 --chmod=777 /tmp /tmp
COPY --from=build --chown=65532:65532 --chmod=644 /docker-entrypoint-initdb.d/ /docker-entrypoint-initdb.d/
COPY --from=build --chown=65532:65532 --chmod=777 /src/docker-postgres/15/alpine/docker-entrypoint.sh /usr/local/bin/entry.sh
RUN chmod -R 777 /tmp; chmod 777 /var/lib/postgresql

USER nonroot
STOPSIGNAL SIGINT

EXPOSE 5432
ENTRYPOINT ["/usr/local/bin/entry.sh"]
CMD ["postgres", "-c", "listen_addresses=*"]
