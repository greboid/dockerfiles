# Generated from https://github.com/greboid/dockerfiles/blob/master/apache/Dockerfile.gotpl
# BOM: {"apk:autoconf":"2.71-r1","apk:bash":"5.2.15-r0","apk:binutils":"2.39-r2","apk:brotli-libs":"1.0.9-r9","apk:build-base":"0.5-r3","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r3","apk:expat":"2.5.0-r0","apk:expat-dev":"2.5.0-r0","apk:expat-static":"2.5.0-r0","apk:file":"5.43-r0","apk:fortify-headers":"1.1-r1","apk:g++":"12.2.1_git20220924-r4","apk:gcc":"12.2.1_git20220924-r4","apk:gdbm":"1.23-r0","apk:git":"2.38.2-r0","apk:gmp":"6.2.1-r2","apk:isl24":"0.24-r1","apk:libacl":"2.3.1-r1","apk:libatomic":"12.2.1_git20220924-r4","apk:libbz2":"1.0.8-r4","apk:libc-dev":"0.7.2-r3","apk:libcrypto3":"3.0.7-r0","apk:libcurl":"7.87.0-r0","apk:libexpat":"2.5.0-r0","apk:libffi":"3.4.4-r0","apk:libgcc":"12.2.1_git20220924-r4","apk:libgomp":"12.2.1_git20220924-r4","apk:libltdl":"2.4.7-r0","apk:libmagic":"5.43-r0","apk:libpcre16":"8.45-r2","apk:libpcre32":"8.45-r2","apk:libpcrecpp":"8.45-r2","apk:libssl3":"3.0.7-r0","apk:libstdc++":"12.2.1_git20220924-r4","apk:libstdc++-dev":"12.2.1_git20220924-r4","apk:libtool":"2.4.7-r0","apk:linux-headers":"5.19.5-r0","apk:lz4-libs":"1.9.4-r1","apk:m4":"1.4.19-r1","apk:make":"4.3-r1","apk:mpc1":"1.2.1-r1","apk:mpdecimal":"2.5.1-r1","apk:mpfr4":"4.1.0-r0","apk:musl":"1.2.3-r4","apk:musl-dev":"1.2.3-r4","apk:ncurses-libs":"6.3_p20221119-r0","apk:ncurses-terminfo-base":"6.3_p20221119-r0","apk:nghttp2-libs":"1.51.0-r0","apk:openssl":"3.0.7-r0","apk:openssl-dev":"3.0.7-r0","apk:openssl-libs-static":"3.0.7-r0","apk:patch":"2.7.6-r8","apk:pcre":"8.45-r2","apk:pcre-dev":"8.45-r2","apk:pcre2":"10.42-r0","apk:perl":"5.36.0-r0","apk:pkgconf":"1.9.3-r0","apk:popt":"1.19-r0","apk:python3":"3.10.9-r1","apk:readline":"8.2.0-r0","apk:rsync":"3.2.7-r0","apk:sqlite-libs":"3.40.0-r0","apk:tzdata":"2022f-r1","apk:xz-libs":"5.2.9-r0","apk:zlib":"1.2.13-r0","apk:zlib-dev":"1.2.13-r0","apk:zlib-static":"1.2.13-r0","apk:zstd-libs":"1.5.2-r9","github:apache/apr":"1.7.0","github:apache/apr-util":"1.6.1","github:apache/httpd":"2.4.54","image:alpine":"3a97af24094652229c76b9dd8dd2895cf50ea52a22b1725db085070041e016ce","image:base":"ffd04efedafc54f5392c35a1045604c12a4329e5780407d5fe08dfe495d411b1"}

FROM reg.g5d.dev/alpine@sha256:3a97af24094652229c76b9dd8dd2895cf50ea52a22b1725db085070041e016ce AS build

ARG APRTAG="1.7.0"
ARG APRUTILTAG="1.6.1"
ARG TAG="2.4.54"

COPY apr-build.patch /apr-build.patch
COPY servertokensoff.patch /servertokensoff.patch

RUN set -eux; \
    apk add --no-cache \
      autoconf=2.71-r1 \
      binutils=2.39-r2 \
      brotli-libs=1.0.9-r9 \
      build-base=0.5-r3 \
      busybox=1.35.0-r29 \
      busybox-binsh=1.35.0-r29 \
      ca-certificates=20220614-r3 \
      expat=2.5.0-r0 \
      expat-dev=2.5.0-r0 \
      expat-static=2.5.0-r0 \
      file=5.43-r0 \
      fortify-headers=1.1-r1 \
      g++=12.2.1_git20220924-r4 \
      gcc=12.2.1_git20220924-r4 \
      gdbm=1.23-r0 \
      git=2.38.2-r0 \
      gmp=6.2.1-r2 \
      isl24=0.24-r1 \
      libacl=2.3.1-r1 \
      libatomic=12.2.1_git20220924-r4 \
      libbz2=1.0.8-r4 \
      libc-dev=0.7.2-r3 \
      libcrypto3=3.0.7-r0 \
      libcurl=7.87.0-r0 \
      libexpat=2.5.0-r0 \
      libffi=3.4.4-r0 \
      libgcc=12.2.1_git20220924-r4 \
      libgomp=12.2.1_git20220924-r4 \
      libltdl=2.4.7-r0 \
      libmagic=5.43-r0 \
      libpcre16=8.45-r2 \
      libpcre32=8.45-r2 \
      libpcrecpp=8.45-r2 \
      libssl3=3.0.7-r0 \
      libstdc++=12.2.1_git20220924-r4 \
      libstdc++-dev=12.2.1_git20220924-r4 \
      libtool=2.4.7-r0 \
      linux-headers=5.19.5-r0 \
      lz4-libs=1.9.4-r1 \
      m4=1.4.19-r1 \
      make=4.3-r1 \
      mpc1=1.2.1-r1 \
      mpdecimal=2.5.1-r1 \
      mpfr4=4.1.0-r0 \
      musl=1.2.3-r4 \
      musl-dev=1.2.3-r4 \
      ncurses-libs=6.3_p20221119-r0 \
      ncurses-terminfo-base=6.3_p20221119-r0 \
      nghttp2-libs=1.51.0-r0 \
      openssl-dev=3.0.7-r0 \
      openssl-libs-static=3.0.7-r0 \
      patch=2.7.6-r8 \
      pcre=8.45-r2 \
      pcre-dev=8.45-r2 \
      pcre2=10.42-r0 \
      perl=5.36.0-r0 \
      pkgconf=1.9.3-r0 \
      popt=1.19-r0 \
      python3=3.10.9-r1 \
      readline=8.2.0-r0 \
      rsync=3.2.7-r0 \
      sqlite-libs=3.40.0-r0 \
      tzdata=2022f-r1 \
      xz-libs=5.2.9-r0 \
      zlib=1.2.13-r0 \
      zlib-dev=1.2.13-r0 \
      zlib-static=1.2.13-r0 \
      zstd-libs=1.5.2-r9 \
      ; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/apache/httpd /src/httpd; \
    git clone --depth=1 -b $APRTAG --single-branch https://github.com/apache/apr /src/httpd/srclib/apr; \
    git clone --depth=1 -b $APRUTILTAG --single-branch https://github.com/apache/apr-util /src/httpd/srclib/apr-util; \
    cd /src/httpd/srclib/apr; \
    git apply /apr-build.patch; \
    cd /src/httpd; \
    git apply /servertokensoff.patch; \
    ./buildconf; \
    cd /src/httpd; \
    ./configure --enable-mpms-shared=all; \ 
    make; \
    make install; \

    apk add bash=5.2.15-r0; \
    apk info -qL bash | rsync -aq --files-from=- / /rootfs/; \
    apk add busybox=1.35.0-r29; \
    apk info -qL busybox | rsync -aq --files-from=- / /rootfs/; \
    apk add busybox-binsh=1.35.0-r29; \
    apk info -qL busybox-binsh | rsync -aq --files-from=- / /rootfs/; \
    apk add ca-certificates=20220614-r3; \
    apk info -qL ca-certificates | rsync -aq --files-from=- / /rootfs/; \
    apk add expat=2.5.0-r0; \
    apk info -qL expat | rsync -aq --files-from=- / /rootfs/; \
    apk add libcrypto3=3.0.7-r0; \
    apk info -qL libcrypto3 | rsync -aq --files-from=- / /rootfs/; \
    apk add libexpat=2.5.0-r0; \
    apk info -qL libexpat | rsync -aq --files-from=- / /rootfs/; \
    apk add libssl3=3.0.7-r0; \
    apk info -qL libssl3 | rsync -aq --files-from=- / /rootfs/; \
    apk add musl=1.2.3-r4; \
    apk info -qL musl | rsync -aq --files-from=- / /rootfs/; \
    apk add ncurses-libs=6.3_p20221119-r0; \
    apk info -qL ncurses-libs | rsync -aq --files-from=- / /rootfs/; \
    apk add ncurses-terminfo-base=6.3_p20221119-r0; \
    apk info -qL ncurses-terminfo-base | rsync -aq --files-from=- / /rootfs/; \
    apk add openssl=3.0.7-r0; \
    apk info -qL openssl | rsync -aq --files-from=- / /rootfs/; \
    apk add pcre=8.45-r2; \
    apk info -qL pcre | rsync -aq --files-from=- / /rootfs/; \
    apk add readline=8.2.0-r0; \
    apk info -qL readline | rsync -aq --files-from=- / /rootfs/; \
    apk add tzdata=2022f-r1; \
    apk info -qL tzdata | rsync -aq --files-from=- / /rootfs/; \
    apk add zlib=1.2.13-r0; \
    apk info -qL zlib | rsync -aq --files-from=- / /rootfs/; \
     \
    mkdir -p /rootfs/usr/local/apache2/bin /rootfs/usr/local/apache2/lib /rootfs/usr/local/apache2/modules /rootfs/usr/local/apache2/logs /rootfs/usr/local/apache2/conf; \
    cp /usr/local/apache2/bin/httpd /rootfs/usr/local/apache2/bin/httpd; \
    cp /usr/local/apache2/conf/mime.types /rootfs/usr/local/apache2/conf/mime.types; \
    rsync -aq /usr/local/apache2/lib/ /rootfs/usr/local/apache2/lib/; \
    rsync -aq /usr/local/apache2/modules/ /rootfs/usr/local/apache2/modules/; \
    rsync -aq /usr/local/apache2/logs/ /rootfs/usr/local/apache2/logs/; \
    chown -R 65532:65532 /rootfs/usr/local/apache2

FROM reg.g5d.dev/base@sha256:ffd04efedafc54f5392c35a1045604c12a4329e5780407d5fe08dfe495d411b1

COPY --from=build /rootfs/ /

ENTRYPOINT ["/usr/local/apache2/bin/httpd", "-DFOREGROUND"]
