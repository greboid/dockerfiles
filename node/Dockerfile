# Generated from https://github.com/csmith/dockerfiles/blob/master/node/Dockerfile.gotpl
# BOM: {"apk:bash":"5.1.4-r0","apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.33.1-r3","apk:ca-certificates":"20191127-r5","apk:curl":"7.79.1-r0","apk:gcompat":"1.0.0-r2","apk:gdbm":"1.19-r0","apk:glib":"2.68.3-r0","apk:gmp":"6.2.1-r0","apk:gnupg":"2.2.31-r0","apk:gnutls":"3.7.1-r0","apk:libacl":"2.2.53-r0","apk:libassuan":"2.5.5-r0","apk:libblkid":"2.37-r0","apk:libbz2":"1.0.8-r1","apk:libcap":"2.50-r0","apk:libcrypto1.1":"1.1.1l-r0","apk:libcurl":"7.79.1-r0","apk:libffi":"3.3-r2","apk:libgcc":"10.3.1_git20210424-r2","apk:libgcrypt":"1.9.4-r0","apk:libgpg-error":"1.42-r0","apk:libintl":"0.21-r0","apk:libksba":"1.5.1-r0","apk:libldap":"2.4.58-r0","apk:libmount":"2.37-r0","apk:libsasl":"2.1.27-r12","apk:libsecret":"0.20.4-r1","apk:libssl1.1":"1.1.1l-r0","apk:libstdc++":"10.3.1_git20210424-r2","apk:libtasn1":"4.17.0-r0","apk:libucontext":"1.1-r0","apk:libunistring":"0.9.10-r1","apk:musl":"1.2.2-r3","apk:musl-obstack":"1.2.2-r0","apk:ncurses-libs":"6.2_p20210612-r0","apk:ncurses-terminfo-base":"6.2_p20210612-r0","apk:nettle":"3.7.3-r0","apk:nghttp2-libs":"1.43.0-r0","apk:npth":"1.6-r0","apk:p11-kit":"0.23.22-r0","apk:pcre":"8.44-r0","apk:pinentry":"1.1.1-r0","apk:popt":"1.18-r0","apk:readline":"8.1.0-r0","apk:rsync":"3.2.3-r4","apk:sqlite-libs":"3.35.5-r0","apk:tar":"1.34-r0","apk:xz":"5.2.5-r0","apk:xz-libs":"5.2.5-r0","apk:zlib":"1.2.11-r3","apk:zstd-libs":"1.4.9-r1","github:nodejs/node":"v16.11.1","image:alpine":"sha256:b81c50e56d1dd83921e2dbf786ef233ac11bf297a9130308f5e1da9f549d7ff3"}

FROM reg.g5d.dev/alpine@sha256:b81c50e56d1dd83921e2dbf786ef233ac11bf297a9130308f5e1da9f549d7ff3 as build

ARG TAG="v16.11.1"

RUN set -eux; \
    apk add --no-cache \
        bash=5.1.4-r0\
        brotli-libs=1.0.9-r5\
        busybox=1.33.1-r3\
        ca-certificates=20191127-r5\
        curl=7.79.1-r0\
        gcompat=1.0.0-r2\
        gdbm=1.19-r0\
        glib=2.68.3-r0\
        gmp=6.2.1-r0\
        gnupg=2.2.31-r0\
        gnutls=3.7.1-r0\
        libacl=2.2.53-r0\
        libassuan=2.5.5-r0\
        libblkid=2.37-r0\
        libbz2=1.0.8-r1\
        libcap=2.50-r0\
        libcrypto1.1=1.1.1l-r0\
        libcurl=7.79.1-r0\
        libffi=3.3-r2\
        libgcc=10.3.1_git20210424-r2\
        libgcrypt=1.9.4-r0\
        libgpg-error=1.42-r0\
        libintl=0.21-r0\
        libksba=1.5.1-r0\
        libldap=2.4.58-r0\
        libmount=2.37-r0\
        libsasl=2.1.27-r12\
        libsecret=0.20.4-r1\
        libssl1.1=1.1.1l-r0\
        libstdc++=10.3.1_git20210424-r2\
        libtasn1=4.17.0-r0\
        libucontext=1.1-r0\
        libunistring=0.9.10-r1\
        musl=1.2.2-r3\
        musl-obstack=1.2.2-r0\
        ncurses-libs=6.2_p20210612-r0\
        ncurses-terminfo-base=6.2_p20210612-r0\
        nettle=3.7.3-r0\
        nghttp2-libs=1.43.0-r0\
        npth=1.6-r0\
        p11-kit=0.23.22-r0\
        pcre=8.44-r0\
        pinentry=1.1.1-r0\
        popt=1.18-r0\
        readline=8.1.0-r0\
        rsync=3.2.3-r4\
        sqlite-libs=3.35.5-r0\
        tar=1.34-r0\
        xz=5.2.5-r0\
        xz-libs=5.2.5-r0\
        zlib=1.2.11-r3\
        zstd-libs=1.4.9-r1\
        ;\
    curl -sfLo /checksums.txt.asc https://nodejs.org/dist/$TAG/SHASUMS256.txt.asc; \
    gpg --keyserver hkps://keys.openpgp.org --keyserver-options auto-key-retrieve --verify /checksums.txt.asc; \
    curl -sfLo /node.tar.xz https://nodejs.org/dist/$TAG/node-$TAG-linux-x64.tar.xz; \
    grep ".*linux-x64.tar.xz" checksums.txt.asc | echo $(awk '{print $1}') \*node.tar.xz | sha256sum -wc -; \
    mkdir -p /src; \
    mkdir -p /usr/local; \
    tar -xf /node.tar.xz -C /src; \
    rsync -ap /src/node*/ /usr/local/; \
    rm -rf /usr/local/CHANGELOG.md; \
    rm -rf /usr/local/LICENSE; \
    rm -rf /usr/local/README.md;

FROM reg.g5d.dev/alpine@sha256:b81c50e56d1dd83921e2dbf786ef233ac11bf297a9130308f5e1da9f549d7ff3

RUN set -eux; \
    apk add --no-cache \
        gcompat=1.0.0-r2\
        libgcc=10.3.1_git20210424-r2\
        libstdc++=10.3.1_git20210424-r2\
        libucontext=1.1-r0\
        musl=1.2.2-r3\
        musl-obstack=1.2.2-r0\
        ;

COPY --from=build /usr/local /usr/local

CMD ["/usr/local/bin/node"]
