# Generated from https://github.com/greboid/dockerfiles/blob/master/texlive/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.34.1-r4","apk:cairo":"1.16.0-r3","apk:expat":"2.4.3-r0","apk:fontconfig":"2.13.1-r4","apk:freetype":"2.11.0-r0","apk:glib":"2.70.1-r0","apk:gmp":"6.2.1-r1","apk:graphite2":"1.3.14-r0","apk:harfbuzz":"3.0.0-r2","apk:icu-libs":"69.1-r1","apk:libacl":"2.2.53-r0","apk:libblkid":"2.37.2-r1","apk:libbsd":"0.11.3-r1","apk:libbz2":"1.0.8-r1","apk:libffi":"3.4.2-r1","apk:libgcc":"10.3.1_git20211027-r0","apk:libintl":"0.21-r0","apk:libmd":"1.0.3-r0","apk:libmount":"2.37.2-r1","apk:libpaper":"1.1.28-r1","apk:libpng":"1.6.37-r1","apk:libstdc++":"10.3.1_git20211027-r0","apk:libuuid":"2.37.2-r1","apk:libx11":"1.7.2-r0","apk:libxau":"1.0.9-r0","apk:libxcb":"1.14-r2","apk:libxdmcp":"1.1.3-r0","apk:libxext":"1.3.4-r0","apk:libxrender":"0.9.10-r3","apk:lz4-libs":"1.9.3-r1","apk:mpfr4":"4.1.0-r0","apk:musl":"1.2.2-r7","apk:pcre":"8.45-r1","apk:perl":"5.34.0-r1","apk:pixman":"0.40.0-r3","apk:popt":"1.18-r0","apk:rsync":"3.2.3-r5","apk:texlive":"20210325-r5","apk:texlive-xetex":"20210325-r5","apk:texmf-dist":"2021.58710-r2","apk:texmf-dist-latexextra":"2021.58710-r2","apk:zlib":"1.2.11-r3","apk:zstd-libs":"1.5.0-r0","image:alpine":"ef95559f475cdf56e55db71c168f1616d497e2215bbd9f6bc63398036ae546a1"}

FROM reg.g5d.dev/alpine@sha256:ef95559f475cdf56e55db71c168f1616d497e2215bbd9f6bc63398036ae546a1 as build

RUN set -eux; \
   apk add --no-cache \
        brotli-libs=1.0.9-r5 \
        busybox=1.34.1-r4 \
        cairo=1.16.0-r3 \
        expat=2.4.3-r0 \
        fontconfig=2.13.1-r4 \
        freetype=2.11.0-r0 \
        glib=2.70.1-r0 \
        gmp=6.2.1-r1 \
        graphite2=1.3.14-r0 \
        harfbuzz=3.0.0-r2 \
        icu-libs=69.1-r1 \
        libacl=2.2.53-r0 \
        libblkid=2.37.2-r1 \
        libbsd=0.11.3-r1 \
        libbz2=1.0.8-r1 \
        libffi=3.4.2-r1 \
        libgcc=10.3.1_git20211027-r0 \
        libintl=0.21-r0 \
        libmd=1.0.3-r0 \
        libmount=2.37.2-r1 \
        libpaper=1.1.28-r1 \
        libpng=1.6.37-r1 \
        libstdc++=10.3.1_git20211027-r0 \
        libuuid=2.37.2-r1 \
        libx11=1.7.2-r0 \
        libxau=1.0.9-r0 \
        libxcb=1.14-r2 \
        libxdmcp=1.1.3-r0 \
        libxext=1.3.4-r0 \
        libxrender=0.9.10-r3 \
        lz4-libs=1.9.3-r1 \
        mpfr4=4.1.0-r0 \
        musl=1.2.2-r7 \
        pcre=8.45-r1 \
        perl=5.34.0-r1 \
        pixman=0.40.0-r3 \
        popt=1.18-r0 \
        rsync=3.2.3-r5 \
        texlive=20210325-r5 \
        texlive-xetex=20210325-r5 \
        texmf-dist=2021.58710-r2 \
        texmf-dist-latexextra=2021.58710-r2 \
        zlib=1.2.11-r3 \
        zstd-libs=1.5.0-r0 \
        ;

WORKDIR /home/nonroot
