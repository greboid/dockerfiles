# Generated from https://github.com/greboid/dockerfiles/blob/master/nginx/Dockerfile.gotpl
# BOM: {"apk:binutils":"2.39-r2","apk:brotli-libs":"1.0.9-r9","apk:build-base":"0.5-r3","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r2","apk:file":"5.43-r0","apk:fortify-headers":"1.1-r1","apk:g++":"12.2.1_git20220924-r4","apk:gcc":"12.2.1_git20220924-r4","apk:git":"2.38.1-r0","apk:gmp":"6.2.1-r2","apk:isl24":"0.24-r1","apk:libatomic":"12.2.1_git20220924-r4","apk:libc-dev":"0.7.2-r3","apk:libcrypto3":"3.0.7-r0","apk:libcurl":"7.86.0-r1","apk:libexpat":"2.5.0-r0","apk:libgcc":"12.2.1_git20220924-r4","apk:libgomp":"12.2.1_git20220924-r4","apk:libmagic":"5.43-r0","apk:libpcre16":"8.45-r2","apk:libpcre32":"8.45-r2","apk:libpcrecpp":"8.45-r2","apk:libssl3":"3.0.7-r0","apk:libstdc++":"12.2.1_git20220924-r4","apk:libstdc++-dev":"12.2.1_git20220924-r4","apk:linux-headers":"5.19.5-r0","apk:make":"4.3-r1","apk:mpc1":"1.2.1-r1","apk:mpfr4":"4.1.0-r0","apk:musl":"1.2.3-r4","apk:musl-dev":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:openssl-dev":"3.0.7-r0","apk:openssl-libs-static":"3.0.7-r0","apk:patch":"2.7.6-r8","apk:pcre":"8.45-r2","apk:pcre-dev":"8.45-r2","apk:pcre2":"10.40-r0","apk:pkgconf":"1.9.3-r0","apk:tzdata":"2022f-r1","apk:zlib":"1.2.13-r0","apk:zlib-dev":"1.2.13-r0","apk:zlib-static":"1.2.13-r0","github:nginx/nginx":"1.23.2","image:alpine":"29ea114ee328eb42c2d032da36bff94b5e0804c9e9a3573d7c5fa052f14da082","image:base":"9af89525fe357fd63d99e49bba4aa94e6e88a7d436500c3588764112f8862109"}

FROM reg.g5d.dev/alpine@sha256:29ea114ee328eb42c2d032da36bff94b5e0804c9e9a3573d7c5fa052f14da082 AS build

ARG TAG="release-1.23.2"

RUN set -eux; \
     apk add --no-cache --virtual .build-deps \
        binutils=2.39-r2 \
        brotli-libs=1.0.9-r9 \
        build-base=0.5-r3 \
        busybox=1.35.0-r29 \
        busybox-binsh=1.35.0-r29 \
        ca-certificates=20220614-r2 \
        file=5.43-r0 \
        fortify-headers=1.1-r1 \
        g++=12.2.1_git20220924-r4 \
        gcc=12.2.1_git20220924-r4 \
        git=2.38.1-r0 \
        gmp=6.2.1-r2 \
        isl24=0.24-r1 \
        libatomic=12.2.1_git20220924-r4 \
        libc-dev=0.7.2-r3 \
        libcrypto3=3.0.7-r0 \
        libcurl=7.86.0-r1 \
        libexpat=2.5.0-r0 \
        libgcc=12.2.1_git20220924-r4 \
        libgomp=12.2.1_git20220924-r4 \
        libmagic=5.43-r0 \
        libpcre16=8.45-r2 \
        libpcre32=8.45-r2 \
        libpcrecpp=8.45-r2 \
        libssl3=3.0.7-r0 \
        libstdc++=12.2.1_git20220924-r4 \
        libstdc++-dev=12.2.1_git20220924-r4 \
        linux-headers=5.19.5-r0 \
        make=4.3-r1 \
        mpc1=1.2.1-r1 \
        mpfr4=4.1.0-r0 \
        musl=1.2.3-r4 \
        musl-dev=1.2.3-r4 \
        nghttp2-libs=1.51.0-r0 \
        openssl-dev=3.0.7-r0 \
        openssl-libs-static=3.0.7-r0 \
        patch=2.7.6-r8 \
        pcre=8.45-r2 \
        pcre-dev=8.45-r2 \
        pcre2=10.40-r0 \
        pkgconf=1.9.3-r0 \
        tzdata=2022f-r1 \
        zlib=1.2.13-r0 \
        zlib-dev=1.2.13-r0 \
        zlib-static=1.2.13-r0 \
        ; \
     git -c advice.detachedHead=false clone --depth=1 -b $TAG --single-branch https://github.com/nginx/nginx /src/nginx; \
     cd /src/nginx; \
     ./auto/configure \
         --without-http_uwsgi_module \
         --without-http_scgi_module \
         --without-http_fastcgi_module \
         --with-http_ssl_module \
         --with-http_v2_module \
         --with-http_realip_module \
         --with-ld-opt="-static"; \
     make install; \
     apk del --no-network .build-deps; \
     mkdir -p /usr/local/nginx/logs \
              /usr/local/nginx/scgi_temp \
              /usr/local/nginx/client_body_temp \
              /usr/local/nginx/client_body_temp \
              /usr/local/nginx/proxy_temp \
              /usr/local/nginx/fastcgi_temp \
              /usr/local/nginx/uwsgi_temp \
              /usr/local/nginx/scgi_temp; \
      ln -s /dev/stdout /usr/local/nginx/logs/access.log; \
      ln -s /dev/stderr /usr/local/nginx/logs/error.log; \
      mkdir -p /nginx/conf; \
      cp -r /usr/local/nginx/logs /nginx; \
      cp -r /usr/local/nginx/logs /nginx; \
      cp -r /usr/local/nginx/uwsgi_temp /nginx; \
      cp -r /usr/local/nginx/fastcgi_temp /nginx; \
      cp -r /usr/local/nginx/proxy_temp /nginx; \
      cp -r /usr/local/nginx/client_body_temp /nginx; \
      cp -r /usr/local/nginx/conf/mime.types /nginx/conf/mime.types;

FROM reg.g5d.dev/base@sha256:9af89525fe357fd63d99e49bba4aa94e6e88a7d436500c3588764112f8862109

COPY --from=build --chown=65532:65532 /nginx /usr/local/nginx
COPY --from=build --chown=65532:65532 /usr/local/nginx/sbin/nginx /nginx

ENTRYPOINT ["/nginx"]
