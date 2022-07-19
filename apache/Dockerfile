# Generated from https://github.com/greboid/dockerfiles/blob/master/apache/Dockerfile.gotpl
# BOM: {"apk:autoconf":"2.71-r0","apk:bash":"5.1.16-r2","apk:binutils":"2.38-r3","apk:brotli-libs":"1.0.9-r6","apk:build-base":"0.5-r3","apk:busybox":"1.35.0-r15","apk:ca-certificates":"20211220-r0","apk:expat":"2.4.8-r0","apk:expat-dev":"2.4.8-r0","apk:expat-static":"2.4.8-r0","apk:file":"5.41-r0","apk:fortify-headers":"1.1-r1","apk:g++":"11.2.1_git20220219-r2","apk:gcc":"11.2.1_git20220219-r2","apk:gdbm":"1.23-r0","apk:git":"2.36.2-r0","apk:gmp":"6.2.1-r2","apk:isl22":"0.22-r0","apk:libacl":"2.3.1-r0","apk:libatomic":"11.2.1_git20220219-r2","apk:libbz2":"1.0.8-r1","apk:libc-dev":"0.7.2-r3","apk:libcrypto1.1":"1.1.1q-r0","apk:libcurl":"7.83.1-r2","apk:libffi":"3.4.2-r1","apk:libgcc":"11.2.1_git20220219-r2","apk:libgomp":"11.2.1_git20220219-r2","apk:libltdl":"2.4.7-r0","apk:libmagic":"5.41-r0","apk:libpcre16":"8.45-r2","apk:libpcre32":"8.45-r2","apk:libpcrecpp":"8.45-r2","apk:libssl1.1":"1.1.1q-r0","apk:libstdc++":"11.2.1_git20220219-r2","apk:libtool":"2.4.7-r0","apk:linux-headers":"5.16.7-r1","apk:lz4-libs":"1.9.3-r1","apk:m4":"1.4.19-r1","apk:make":"4.3-r0","apk:mpc1":"1.2.1-r0","apk:mpdecimal":"2.5.1-r1","apk:mpfr4":"4.1.0-r0","apk:musl":"1.2.3-r0","apk:musl-dev":"1.2.3-r0","apk:ncurses-libs":"6.3_p20220521-r0","apk:ncurses-terminfo-base":"6.3_p20220521-r0","apk:nghttp2-libs":"1.47.0-r0","apk:openssl":"1.1.1q-r0","apk:openssl-dev":"1.1.1q-r0","apk:openssl-libs-static":"1.1.1q-r0","apk:patch":"2.7.6-r7","apk:pcre":"8.45-r2","apk:pcre-dev":"8.45-r2","apk:pcre2":"10.40-r0","apk:perl":"5.34.1-r0","apk:pkgconf":"1.8.0-r0","apk:popt":"1.18-r0","apk:python3":"3.10.4-r0","apk:pythonispython3":"3.10.4-r0","apk:readline":"8.1.2-r0","apk:rsync":"3.2.4-r1","apk:sqlite-libs":"3.38.5-r0","apk:tzdata":"2022a-r0","apk:xz-libs":"5.2.5-r1","apk:zlib":"1.2.12-r1","apk:zlib-dev":"1.2.12-r1","apk:zlib-static":"1.2.12-r1","apk:zstd-libs":"1.5.2-r1","github:apache/apr":"1.7.0","github:apache/apr-util":"1.6.1","github:apache/httpd":"2.4.54","image:alpine":"9169c79c67da14335d9f81a00e4910cd79d1811f9f44a47cd8e520fc4265e1ae","image:base":"d8e173562b2af5c2d2a6bedc5af4a36a9fe9e9561aca25d22c947752a1695a64"}

FROM reg.g5d.dev/alpine@sha256:9169c79c67da14335d9f81a00e4910cd79d1811f9f44a47cd8e520fc4265e1ae AS build

ARG APRTAG="1.7.0"
ARG APRUTILTAG="1.6.1"
ARG TAG="2.4.54"

COPY apr-build.patch /apr-build.patch
COPY servertokensoff.patch /servertokensoff.patch

RUN set -eux; \
    apk add --no-cache \
      autoconf=2.71-r0 \
      binutils=2.38-r3 \
      brotli-libs=1.0.9-r6 \
      build-base=0.5-r3 \
      busybox=1.35.0-r15 \
      ca-certificates=20211220-r0 \
      expat=2.4.8-r0 \
      expat-dev=2.4.8-r0 \
      expat-static=2.4.8-r0 \
      file=5.41-r0 \
      fortify-headers=1.1-r1 \
      g++=11.2.1_git20220219-r2 \
      gcc=11.2.1_git20220219-r2 \
      gdbm=1.23-r0 \
      git=2.36.2-r0 \
      gmp=6.2.1-r2 \
      isl22=0.22-r0 \
      libacl=2.3.1-r0 \
      libatomic=11.2.1_git20220219-r2 \
      libbz2=1.0.8-r1 \
      libc-dev=0.7.2-r3 \
      libcrypto1.1=1.1.1q-r0 \
      libcurl=7.83.1-r2 \
      libffi=3.4.2-r1 \
      libgcc=11.2.1_git20220219-r2 \
      libgomp=11.2.1_git20220219-r2 \
      libltdl=2.4.7-r0 \
      libmagic=5.41-r0 \
      libpcre16=8.45-r2 \
      libpcre32=8.45-r2 \
      libpcrecpp=8.45-r2 \
      libssl1.1=1.1.1q-r0 \
      libstdc++=11.2.1_git20220219-r2 \
      libtool=2.4.7-r0 \
      linux-headers=5.16.7-r1 \
      lz4-libs=1.9.3-r1 \
      m4=1.4.19-r1 \
      make=4.3-r0 \
      mpc1=1.2.1-r0 \
      mpdecimal=2.5.1-r1 \
      mpfr4=4.1.0-r0 \
      musl=1.2.3-r0 \
      musl-dev=1.2.3-r0 \
      ncurses-libs=6.3_p20220521-r0 \
      ncurses-terminfo-base=6.3_p20220521-r0 \
      nghttp2-libs=1.47.0-r0 \
      openssl-dev=1.1.1q-r0 \
      openssl-libs-static=1.1.1q-r0 \
      patch=2.7.6-r7 \
      pcre=8.45-r2 \
      pcre-dev=8.45-r2 \
      pcre2=10.40-r0 \
      perl=5.34.1-r0 \
      pkgconf=1.8.0-r0 \
      popt=1.18-r0 \
      python3=3.10.4-r0 \
      pythonispython3=3.10.4-r0 \
      readline=8.1.2-r0 \
      rsync=3.2.4-r1 \
      sqlite-libs=3.38.5-r0 \
      tzdata=2022a-r0 \
      xz-libs=5.2.5-r1 \
      zlib=1.2.12-r1 \
      zlib-dev=1.2.12-r1 \
      zlib-static=1.2.12-r1 \
      zstd-libs=1.5.2-r1 \
      ; \
    git -c advice.detachedHead=false clone --depth=1 -b $TAG --single-branch https://github.com/apache/httpd /src/httpd; \
    git -c advice.detachedHead=false clone --depth=1 -b $APRTAG --single-branch https://github.com/apache/apr /src/httpd/srclib/apr; \
    git -c advice.detachedHead=false clone --depth=1 -b $APRUTILTAG --single-branch https://github.com/apache/apr-util /src/httpd/srclib/apr-util; \
    cd /src/httpd/srclib/apr; \
    git apply /apr-build.patch; \
    cd /src/httpd; \
    git apply /servertokensoff.patch; \
    ./buildconf; \
    cd /src/httpd; \
    ./configure --enable-mpms-shared=all; \ 
    make; \
    make install; \

    apk add bash=5.1.16-r2; \
    apk info -qL bash | rsync -aq --files-from=- / /rootfs/; \
    apk add busybox=1.35.0-r15; \
    apk info -qL busybox | rsync -aq --files-from=- / /rootfs/; \
    apk add ca-certificates=20211220-r0; \
    apk info -qL ca-certificates | rsync -aq --files-from=- / /rootfs/; \
    apk add expat=2.4.8-r0; \
    apk info -qL expat | rsync -aq --files-from=- / /rootfs/; \
    apk add libcrypto1.1=1.1.1q-r0; \
    apk info -qL libcrypto1.1 | rsync -aq --files-from=- / /rootfs/; \
    apk add libssl1.1=1.1.1q-r0; \
    apk info -qL libssl1.1 | rsync -aq --files-from=- / /rootfs/; \
    apk add musl=1.2.3-r0; \
    apk info -qL musl | rsync -aq --files-from=- / /rootfs/; \
    apk add ncurses-libs=6.3_p20220521-r0; \
    apk info -qL ncurses-libs | rsync -aq --files-from=- / /rootfs/; \
    apk add ncurses-terminfo-base=6.3_p20220521-r0; \
    apk info -qL ncurses-terminfo-base | rsync -aq --files-from=- / /rootfs/; \
    apk add openssl=1.1.1q-r0; \
    apk info -qL openssl | rsync -aq --files-from=- / /rootfs/; \
    apk add pcre=8.45-r2; \
    apk info -qL pcre | rsync -aq --files-from=- / /rootfs/; \
    apk add readline=8.1.2-r0; \
    apk info -qL readline | rsync -aq --files-from=- / /rootfs/; \
    apk add tzdata=2022a-r0; \
    apk info -qL tzdata | rsync -aq --files-from=- / /rootfs/; \
    apk add zlib=1.2.12-r1; \
    apk info -qL zlib | rsync -aq --files-from=- / /rootfs/; \
     \
    mkdir -p /rootfs/usr/local/apache2/bin /rootfs/usr/local/apache2/lib /rootfs/usr/local/apache2/modules /rootfs/usr/local/apache2/logs /rootfs/usr/local/apache2/conf; \
    cp /usr/local/apache2/bin/httpd /rootfs/usr/local/apache2/bin/httpd; \
    cp /usr/local/apache2/conf/mime.types /rootfs/usr/local/apache2/conf/mime.types; \
    rsync -aq /usr/local/apache2/lib/ /rootfs/usr/local/apache2/lib/; \
    rsync -aq /usr/local/apache2/modules/ /rootfs/usr/local/apache2/modules/; \
    rsync -aq /usr/local/apache2/logs/ /rootfs/usr/local/apache2/logs/; \
    chown -R 65532:65532 /rootfs/usr/local/apache2

FROM reg.g5d.dev/base@sha256:d8e173562b2af5c2d2a6bedc5af4a36a9fe9e9561aca25d22c947752a1695a64

COPY --from=build /rootfs/ /

ENTRYPOINT ["/usr/local/apache2/bin/httpd", "-DFOREGROUND"]
