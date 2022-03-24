# Generated from https://github.com/greboid/dockerfiles/blob/master/nginx/Dockerfile.gotpl
# BOM: {"apk:binutils":"2.37-r3","apk:brotli-libs":"1.0.9-r5","apk:build-base":"0.5-r2","apk:busybox":"1.34.1-r4","apk:ca-certificates":"20211220-r0","apk:expat":"2.4.7-r0","apk:file":"5.41-r0","apk:fortify-headers":"1.1-r1","apk:g++":"10.3.1_git20211027-r0","apk:gcc":"10.3.1_git20211027-r0","apk:git":"2.34.1-r0","apk:gmp":"6.2.1-r1","apk:isl22":"0.22-r0","apk:libatomic":"10.3.1_git20211027-r0","apk:libc-dev":"0.7.2-r3","apk:libcrypto1.1":"1.1.1n-r0","apk:libcurl":"7.80.0-r0","apk:libgcc":"10.3.1_git20211027-r0","apk:libgomp":"10.3.1_git20211027-r0","apk:libgphobos":"10.3.1_git20211027-r0","apk:libmagic":"5.41-r0","apk:libpcre16":"8.45-r1","apk:libpcre32":"8.45-r1","apk:libpcrecpp":"8.45-r1","apk:libssl1.1":"1.1.1n-r0","apk:libstdc++":"10.3.1_git20211027-r0","apk:linux-headers":"5.10.41-r0","apk:make":"4.3-r0","apk:mpc1":"1.2.1-r0","apk:mpfr4":"4.1.0-r0","apk:musl":"1.2.2-r7","apk:musl-dev":"1.2.2-r7","apk:nghttp2-libs":"1.46.0-r0","apk:openssl-dev":"1.1.1n-r0","apk:openssl-libs-static":"1.1.1n-r0","apk:patch":"2.7.6-r7","apk:pcre":"8.45-r1","apk:pcre-dev":"8.45-r1","apk:pcre2":"10.39-r0","apk:pkgconf":"1.8.0-r0","apk:tzdata":"2022a-r0","apk:zlib":"1.2.11-r3","apk:zlib-dev":"1.2.11-r3","apk:zlib-static":"1.2.11-r3","github:nginx/nginx":"1.21.6","image:alpine":"a4460332f57a36a82ca308dd52b6565829a025b66580248508b45908f0061bc1","image:base":"ee8879a758bbe46a6b0b5b9d51a3cd3cb5660106d185b0f1fd981e4475663d08"}

FROM reg.g5d.dev/alpine@sha256:a4460332f57a36a82ca308dd52b6565829a025b66580248508b45908f0061bc1 AS build

ARG TAG="release-1.21.6"

RUN set -eux; \
     apk add --no-cache --virtual .build-deps \
        binutils=2.37-r3 \
        brotli-libs=1.0.9-r5 \
        build-base=0.5-r2 \
        busybox=1.34.1-r4 \
        ca-certificates=20211220-r0 \
        expat=2.4.7-r0 \
        file=5.41-r0 \
        fortify-headers=1.1-r1 \
        g++=10.3.1_git20211027-r0 \
        gcc=10.3.1_git20211027-r0 \
        git=2.34.1-r0 \
        gmp=6.2.1-r1 \
        isl22=0.22-r0 \
        libatomic=10.3.1_git20211027-r0 \
        libc-dev=0.7.2-r3 \
        libcrypto1.1=1.1.1n-r0 \
        libcurl=7.80.0-r0 \
        libgcc=10.3.1_git20211027-r0 \
        libgomp=10.3.1_git20211027-r0 \
        libgphobos=10.3.1_git20211027-r0 \
        libmagic=5.41-r0 \
        libpcre16=8.45-r1 \
        libpcre32=8.45-r1 \
        libpcrecpp=8.45-r1 \
        libssl1.1=1.1.1n-r0 \
        libstdc++=10.3.1_git20211027-r0 \
        linux-headers=5.10.41-r0 \
        make=4.3-r0 \
        mpc1=1.2.1-r0 \
        mpfr4=4.1.0-r0 \
        musl=1.2.2-r7 \
        musl-dev=1.2.2-r7 \
        nghttp2-libs=1.46.0-r0 \
        openssl-dev=1.1.1n-r0 \
        openssl-libs-static=1.1.1n-r0 \
        patch=2.7.6-r7 \
        pcre=8.45-r1 \
        pcre-dev=8.45-r1 \
        pcre2=10.39-r0 \
        pkgconf=1.8.0-r0 \
        tzdata=2022a-r0 \
        zlib=1.2.11-r3 \
        zlib-dev=1.2.11-r3 \
        zlib-static=1.2.11-r3 \
        ; \
     git -c advice.detachedHead=false clone --depth=1 -b $TAG --single-branch https://github.com/nginx/nginx /src/nginx; \
     cd /src/nginx; \
     ./auto/configure \
         --without-http_uwsgi_module \
         --without-http_scgi_module \
         --without-http_fastcgi_module \
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

FROM reg.g5d.dev/base@sha256:ee8879a758bbe46a6b0b5b9d51a3cd3cb5660106d185b0f1fd981e4475663d08

COPY --from=build --chown=65532:65532 /nginx /usr/local/nginx
COPY --from=build --chown=65532:65532 /usr/local/nginx/sbin/nginx /nginx

ENTRYPOINT ["/nginx"]
