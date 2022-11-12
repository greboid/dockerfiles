# Generated from https://github.com/greboid/dockerfiles/blob/master/nginx/Dockerfile.gotpl
# BOM: {"apk:binutils":"2.38-r3","apk:brotli-libs":"1.0.9-r6","apk:build-base":"0.5-r3","apk:busybox":"1.35.0-r17","apk:ca-certificates":"20220614-r0","apk:expat":"2.5.0-r0","apk:file":"5.41-r0","apk:fortify-headers":"1.1-r1","apk:g++":"11.2.1_git20220219-r2","apk:gcc":"11.2.1_git20220219-r2","apk:git":"2.36.3-r0","apk:gmp":"6.2.1-r2","apk:isl22":"0.22-r0","apk:libatomic":"11.2.1_git20220219-r2","apk:libc-dev":"0.7.2-r3","apk:libcrypto1.1":"1.1.1s-r0","apk:libcurl":"7.83.1-r4","apk:libgcc":"11.2.1_git20220219-r2","apk:libgomp":"11.2.1_git20220219-r2","apk:libmagic":"5.41-r0","apk:libpcre16":"8.45-r2","apk:libpcre32":"8.45-r2","apk:libpcrecpp":"8.45-r2","apk:libssl1.1":"1.1.1s-r0","apk:libstdc++":"11.2.1_git20220219-r2","apk:linux-headers":"5.16.7-r1","apk:make":"4.3-r0","apk:mpc1":"1.2.1-r0","apk:mpfr4":"4.1.0-r0","apk:musl":"1.2.3-r2","apk:musl-dev":"1.2.3-r2","apk:nghttp2-libs":"1.47.0-r0","apk:openssl-dev":"1.1.1s-r0","apk:openssl-libs-static":"1.1.1s-r0","apk:patch":"2.7.6-r7","apk:pcre":"8.45-r2","apk:pcre-dev":"8.45-r2","apk:pcre2":"10.40-r0","apk:pkgconf":"1.8.0-r1","apk:tzdata":"2022f-r1","apk:zlib":"1.2.12-r3","apk:zlib-dev":"1.2.12-r3","apk:zlib-static":"1.2.12-r3","github:nginx/nginx":"1.23.2","image:alpine":"1f6a4275acd73275186069c89bdb401e7dff76ebcd2fe8469a10d4b698d34512","image:base":"94a9a1a6a1fa72a3d617c43e69f1a125e82e49d5381b3ea37dfaafc2a21ef3ee"}

FROM reg.g5d.dev/alpine@sha256:1f6a4275acd73275186069c89bdb401e7dff76ebcd2fe8469a10d4b698d34512 AS build

ARG TAG="release-1.23.2"

RUN set -eux; \
     apk add --no-cache --virtual .build-deps \
        binutils=2.38-r3 \
        brotli-libs=1.0.9-r6 \
        build-base=0.5-r3 \
        busybox=1.35.0-r17 \
        ca-certificates=20220614-r0 \
        expat=2.5.0-r0 \
        file=5.41-r0 \
        fortify-headers=1.1-r1 \
        g++=11.2.1_git20220219-r2 \
        gcc=11.2.1_git20220219-r2 \
        git=2.36.3-r0 \
        gmp=6.2.1-r2 \
        isl22=0.22-r0 \
        libatomic=11.2.1_git20220219-r2 \
        libc-dev=0.7.2-r3 \
        libcrypto1.1=1.1.1s-r0 \
        libcurl=7.83.1-r4 \
        libgcc=11.2.1_git20220219-r2 \
        libgomp=11.2.1_git20220219-r2 \
        libmagic=5.41-r0 \
        libpcre16=8.45-r2 \
        libpcre32=8.45-r2 \
        libpcrecpp=8.45-r2 \
        libssl1.1=1.1.1s-r0 \
        libstdc++=11.2.1_git20220219-r2 \
        linux-headers=5.16.7-r1 \
        make=4.3-r0 \
        mpc1=1.2.1-r0 \
        mpfr4=4.1.0-r0 \
        musl=1.2.3-r2 \
        musl-dev=1.2.3-r2 \
        nghttp2-libs=1.47.0-r0 \
        openssl-dev=1.1.1s-r0 \
        openssl-libs-static=1.1.1s-r0 \
        patch=2.7.6-r7 \
        pcre=8.45-r2 \
        pcre-dev=8.45-r2 \
        pcre2=10.40-r0 \
        pkgconf=1.8.0-r1 \
        tzdata=2022f-r1 \
        zlib=1.2.12-r3 \
        zlib-dev=1.2.12-r3 \
        zlib-static=1.2.12-r3 \
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

FROM reg.g5d.dev/base@sha256:94a9a1a6a1fa72a3d617c43e69f1a125e82e49d5381b3ea37dfaafc2a21ef3ee

COPY --from=build --chown=65532:65532 /nginx /usr/local/nginx
COPY --from=build --chown=65532:65532 /usr/local/nginx/sbin/nginx /nginx

ENTRYPOINT ["/nginx"]
