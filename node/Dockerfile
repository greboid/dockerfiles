# Generated from https://github.com/greboid/dockerfiles/blob/master/node/Dockerfile.gotpl
# BOM: {"apk:bash":"5.2.9-r0","apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r2","apk:curl":"7.86.0-r1","apk:gcompat":"1.1.0-r0","apk:gdbm":"1.23-r0","apk:git":"2.38.1-r0","apk:gmp":"6.2.1-r2","apk:gnupg":"2.2.40-r0","apk:gnupg-dirmngr":"2.2.40-r0","apk:gnupg-gpgconf":"2.2.40-r0","apk:gnupg-utils":"2.2.40-r0","apk:gnupg-wks-client":"2.2.40-r0","apk:gnutls":"3.7.8-r2","apk:gpg":"2.2.40-r0","apk:gpg-agent":"2.2.40-r0","apk:gpg-wks-server":"2.2.40-r0","apk:gpgsm":"2.2.40-r0","apk:gpgv":"2.2.40-r0","apk:libacl":"2.3.1-r1","apk:libassuan":"2.5.5-r1","apk:libbz2":"1.0.8-r4","apk:libcrypto3":"3.0.7-r0","apk:libcurl":"7.86.0-r1","apk:libexpat":"2.5.0-r0","apk:libffi":"3.4.4-r0","apk:libgcc":"12.2.1_git20220924-r4","apk:libgcrypt":"1.10.1-r0","apk:libgpg-error":"1.46-r1","apk:libksba":"1.6.2-r0","apk:libldap":"2.6.3-r6","apk:libsasl":"2.1.28-r3","apk:libssl3":"3.0.7-r0","apk:libstdc++":"12.2.1_git20220924-r4","apk:libtasn1":"4.19.0-r0","apk:libucontext":"1.2-r0","apk:libunistring":"1.1-r0","apk:lz4-libs":"1.9.4-r1","apk:musl":"1.2.3-r4","apk:musl-obstack":"1.2.3-r0","apk:ncurses-libs":"6.3_p20221119-r0","apk:ncurses-terminfo-base":"6.3_p20221119-r0","apk:nettle":"3.8.1-r0","apk:nghttp2-libs":"1.51.0-r0","apk:npth":"1.6-r2","apk:p11-kit":"0.24.1-r1","apk:pcre2":"10.40-r0","apk:pinentry":"1.2.1-r0","apk:popt":"1.19-r0","apk:readline":"8.2.0-r0","apk:rsync":"3.2.7-r0","apk:sqlite-libs":"3.40.0-r0","apk:tar":"1.34-r1","apk:xz":"5.2.8-r0","apk:xz-libs":"5.2.8-r0","apk:zlib":"1.2.13-r0","apk:zstd-libs":"1.5.2-r9","github:nodejs/node":"v19.1.0","image:alpine":"29ea114ee328eb42c2d032da36bff94b5e0804c9e9a3573d7c5fa052f14da082"}

FROM reg.g5d.dev/alpine@sha256:29ea114ee328eb42c2d032da36bff94b5e0804c9e9a3573d7c5fa052f14da082 as build

ARG TAG="v19.1.0"

RUN set -eux; \
    apk add --no-cache \
        bash=5.2.9-r0\
        brotli-libs=1.0.9-r9\
        busybox=1.35.0-r29\
        busybox-binsh=1.35.0-r29\
        ca-certificates=20220614-r2\
        curl=7.86.0-r1\
        gcompat=1.1.0-r0\
        gdbm=1.23-r0\
        gmp=6.2.1-r2\
        gnupg=2.2.40-r0\
        gnupg-dirmngr=2.2.40-r0\
        gnupg-gpgconf=2.2.40-r0\
        gnupg-utils=2.2.40-r0\
        gnupg-wks-client=2.2.40-r0\
        gnutls=3.7.8-r2\
        gpg=2.2.40-r0\
        gpg-agent=2.2.40-r0\
        gpg-wks-server=2.2.40-r0\
        gpgsm=2.2.40-r0\
        gpgv=2.2.40-r0\
        libacl=2.3.1-r1\
        libassuan=2.5.5-r1\
        libbz2=1.0.8-r4\
        libcrypto3=3.0.7-r0\
        libcurl=7.86.0-r1\
        libffi=3.4.4-r0\
        libgcc=12.2.1_git20220924-r4\
        libgcrypt=1.10.1-r0\
        libgpg-error=1.46-r1\
        libksba=1.6.2-r0\
        libldap=2.6.3-r6\
        libsasl=2.1.28-r3\
        libssl3=3.0.7-r0\
        libstdc++=12.2.1_git20220924-r4\
        libtasn1=4.19.0-r0\
        libucontext=1.2-r0\
        libunistring=1.1-r0\
        lz4-libs=1.9.4-r1\
        musl=1.2.3-r4\
        musl-obstack=1.2.3-r0\
        ncurses-libs=6.3_p20221119-r0\
        ncurses-terminfo-base=6.3_p20221119-r0\
        nettle=3.8.1-r0\
        nghttp2-libs=1.51.0-r0\
        npth=1.6-r2\
        p11-kit=0.24.1-r1\
        pinentry=1.2.1-r0\
        popt=1.19-r0\
        readline=8.2.0-r0\
        rsync=3.2.7-r0\
        sqlite-libs=3.40.0-r0\
        tar=1.34-r1\
        xz=5.2.8-r0\
        xz-libs=5.2.8-r0\
        zlib=1.2.13-r0\
        zstd-libs=1.5.2-r9\
        ;\
    curl -sfLo /checksums.txt https://unofficial-builds.nodejs.org/download/release/$TAG/SHASUMS256.txt; \
    curl -sfLo /node.tar.xz https://unofficial-builds.nodejs.org/download/release//$TAG/node-$TAG-linux-x64-musl.tar.xz; \
    grep ".*linux-x64-musl.tar.xz" checksums.txt | echo $(awk '{print $1}') \*node.tar.xz | sha256sum -wc -; \
    mkdir -p /src; \
    mkdir -p /usr/local; \
    tar -xf /node.tar.xz -C /src; \
    rsync -ap /src/node*/ /usr/local/; \
    rm -rf /usr/local/CHANGELOG.md; \
    rm -rf /usr/local/LICENSE; \
    rm -rf /usr/local/README.md;

FROM reg.g5d.dev/alpine@sha256:29ea114ee328eb42c2d032da36bff94b5e0804c9e9a3573d7c5fa052f14da082

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r9\
        busybox=1.35.0-r29\
        busybox-binsh=1.35.0-r29\
        ca-certificates=20220614-r2\
        gcompat=1.1.0-r0\
        git=2.38.1-r0\
        libcrypto3=3.0.7-r0\
        libcurl=7.86.0-r1\
        libexpat=2.5.0-r0\
        libgcc=12.2.1_git20220924-r4\
        libssl3=3.0.7-r0\
        libstdc++=12.2.1_git20220924-r4\
        libucontext=1.2-r0\
        musl=1.2.3-r4\
        musl-obstack=1.2.3-r0\
        nghttp2-libs=1.51.0-r0\
        pcre2=10.40-r0\
        zlib=1.2.13-r0\
        ;

COPY --from=build /usr/local /usr/local

CMD ["/usr/local/bin/node"]
