# Generated from https://github.com/greboid/dockerfiles/blob/master/texlive/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.33.1-r3","apk:cairo":"1.16.0-r3","apk:expat":"2.4.1-r0","apk:fontconfig":"2.13.1-r4","apk:freetype":"2.10.4-r1","apk:glib":"2.68.3-r0","apk:gmp":"6.2.1-r0","apk:graphite2":"1.3.14-r0","apk:harfbuzz":"2.8.1-r0","apk:icu-libs":"67.1-r2","apk:libacl":"2.2.53-r0","apk:libblkid":"2.37-r0","apk:libbsd":"0.11.3-r0","apk:libbz2":"1.0.8-r1","apk:libffi":"3.3-r2","apk:libgcc":"10.3.1_git20210424-r2","apk:libintl":"0.21-r0","apk:libmd":"1.0.3-r0","apk:libmount":"2.37-r0","apk:libpaper":"1.1.28-r0","apk:libpng":"1.6.37-r1","apk:libstdc++":"10.3.1_git20210424-r2","apk:libuuid":"2.37-r0","apk:libx11":"1.7.2-r0","apk:libxau":"1.0.9-r0","apk:libxcb":"1.14-r2","apk:libxdmcp":"1.1.3-r0","apk:libxext":"1.3.4-r0","apk:libxrender":"0.9.10-r3","apk:mpfr4":"4.1.0-r0","apk:musl":"1.2.2-r3","apk:pcre":"8.44-r0","apk:perl":"5.32.1-r0","apk:pixman":"0.40.0-r2","apk:popt":"1.18-r0","apk:rsync":"3.2.3-r4","apk:texlive":"20210325-r4","apk:texlive-xetex":"20210325-r4","apk:texmf-dist":"2021.58710-r1","apk:texmf-dist-latexextra":"2021.58710-r1","apk:zlib":"1.2.11-r3","apk:zstd-libs":"1.4.9-r1","image:alpine":"c8797cc5b046929decf47059cc3a62b89d664ed442288d10833f496d1339af7f","image:base":"f25b9dc29d96dee9ff3919d93729ba9105802a4ee8a15b69e654877fa74e0af2"}

FROM reg.g5d.dev/alpine@sha256:c8797cc5b046929decf47059cc3a62b89d664ed442288d10833f496d1339af7f as build

RUN set -eux; \
   apk add --no-cache \
        brotli-libs=1.0.9-r5 \
        busybox=1.33.1-r3 \
        cairo=1.16.0-r3 \
        expat=2.4.1-r0 \
        fontconfig=2.13.1-r4 \
        freetype=2.10.4-r1 \
        glib=2.68.3-r0 \
        gmp=6.2.1-r0 \
        graphite2=1.3.14-r0 \
        harfbuzz=2.8.1-r0 \
        icu-libs=67.1-r2 \
        libacl=2.2.53-r0 \
        libblkid=2.37-r0 \
        libbsd=0.11.3-r0 \
        libbz2=1.0.8-r1 \
        libffi=3.3-r2 \
        libgcc=10.3.1_git20210424-r2 \
        libintl=0.21-r0 \
        libmd=1.0.3-r0 \
        libmount=2.37-r0 \
        libpaper=1.1.28-r0 \
        libpng=1.6.37-r1 \
        libstdc++=10.3.1_git20210424-r2 \
        libuuid=2.37-r0 \
        libx11=1.7.2-r0 \
        libxau=1.0.9-r0 \
        libxcb=1.14-r2 \
        libxdmcp=1.1.3-r0 \
        libxext=1.3.4-r0 \
        libxrender=0.9.10-r3 \
        mpfr4=4.1.0-r0 \
        musl=1.2.2-r3 \
        pcre=8.44-r0 \
        perl=5.32.1-r0 \
        pixman=0.40.0-r2 \
        popt=1.18-r0 \
        rsync=3.2.3-r4 \
        texlive=20210325-r4 \
        texlive-xetex=20210325-r4 \
        texmf-dist=2021.58710-r1 \
        texmf-dist-latexextra=2021.58710-r1 \
        zlib=1.2.11-r3 \
        zstd-libs=1.4.9-r1 \
        ; \
   apk info -qL brotli-libs | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL busybox | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL cairo | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL expat | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL fontconfig | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL freetype | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL glib | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL gmp | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL graphite2 | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL harfbuzz | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL icu-libs | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL libacl | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL libblkid | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL libbsd | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL libbz2 | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL libffi | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL libgcc | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL libintl | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL libmd | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL libmount | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL libpaper | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL libpng | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL libstdc++ | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL libuuid | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL libx11 | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL libxau | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL libxcb | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL libxdmcp | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL libxext | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL libxrender | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL mpfr4 | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL musl | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL pcre | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL perl | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL pixman | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL popt | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL rsync | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL texlive | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL texlive-xetex | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL texmf-dist | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL texmf-dist-latexextra | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL zlib | rsync -aq --files-from=- / /rootfs/; \
   apk info -qL zstd-libs | rsync -aq --files-from=- / /rootfs/; \
    \
   mkdir -p /rootfs/bin; \
   cp /bin/sh /rootfs/bin/sh; \
   find /rootfs -type f -exec touch --date=@0 {} \;

FROM reg.g5d.dev/base@sha256:f25b9dc29d96dee9ff3919d93729ba9105802a4ee8a15b69e654877fa74e0af2

COPY --from=build /rootfs/ /
