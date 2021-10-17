# Generated from https://github.com/csmith/dockerfiles/blob/master/nginx/Dockerfile.gotpl
# BOM: {"apk:binutils":"2.35.2-r2","apk:brotli-libs":"1.0.9-r5","apk:build-base":"0.5-r2","apk:busybox":"1.33.1-r3","apk:ca-certificates":"20191127-r5","apk:expat":"2.4.1-r0","apk:file":"5.40-r1","apk:fortify-headers":"1.1-r1","apk:g++":"10.3.1_git20210424-r2","apk:gcc":"10.3.1_git20210424-r2","apk:git":"2.32.0-r0","apk:gmp":"6.2.1-r0","apk:isl22":"0.22-r0","apk:libatomic":"10.3.1_git20210424-r2","apk:libc-dev":"0.7.2-r3","apk:libcrypto1.1":"1.1.1l-r0","apk:libcurl":"7.79.1-r0","apk:libgcc":"10.3.1_git20210424-r2","apk:libgomp":"10.3.1_git20210424-r2","apk:libgphobos":"10.3.1_git20210424-r2","apk:libmagic":"5.40-r1","apk:libpcre16":"8.44-r0","apk:libpcre32":"8.44-r0","apk:libpcrecpp":"8.44-r0","apk:libssl1.1":"1.1.1l-r0","apk:libstdc++":"10.3.1_git20210424-r2","apk:linux-headers":"5.10.41-r0","apk:make":"4.3-r0","apk:mpc1":"1.2.1-r0","apk:mpfr4":"4.1.0-r0","apk:musl":"1.2.2-r3","apk:musl-dev":"1.2.2-r3","apk:nghttp2-libs":"1.43.0-r0","apk:openssl-dev":"1.1.1l-r0","apk:openssl-libs-static":"1.1.1l-r0","apk:patch":"2.7.6-r7","apk:pcre":"8.44-r0","apk:pcre-dev":"8.44-r0","apk:pcre2":"10.36-r0","apk:pkgconf":"1.7.4-r0","apk:tzdata":"2021d-r0","apk:zlib":"1.2.11-r3","apk:zlib-dev":"1.2.11-r3","apk:zlib-static":"1.2.11-r3","image:alpine":"sha256:b81c50e56d1dd83921e2dbf786ef233ac11bf297a9130308f5e1da9f549d7ff3","image:base":"sha256:bc61849aeec7da9f45293bb26ba1e37240ccb66a96c173ec8cfc817e71e4ce90"}

FROM reg.g5d.dev/alpine@sha256:b81c50e56d1dd83921e2dbf786ef233ac11bf297a9130308f5e1da9f549d7ff3 AS build

ARG TAG="release-1.21.3"

RUN set -eux; \
     apk add --no-cache --virtual .build-deps \
        binutils=2.35.2-r2 \
        brotli-libs=1.0.9-r5 \
        build-base=0.5-r2 \
        busybox=1.33.1-r3 \
        ca-certificates=20191127-r5 \
        expat=2.4.1-r0 \
        file=5.40-r1 \
        fortify-headers=1.1-r1 \
        g++=10.3.1_git20210424-r2 \
        gcc=10.3.1_git20210424-r2 \
        git=2.32.0-r0 \
        gmp=6.2.1-r0 \
        isl22=0.22-r0 \
        libatomic=10.3.1_git20210424-r2 \
        libc-dev=0.7.2-r3 \
        libcrypto1.1=1.1.1l-r0 \
        libcurl=7.79.1-r0 \
        libgcc=10.3.1_git20210424-r2 \
        libgomp=10.3.1_git20210424-r2 \
        libgphobos=10.3.1_git20210424-r2 \
        libmagic=5.40-r1 \
        libpcre16=8.44-r0 \
        libpcre32=8.44-r0 \
        libpcrecpp=8.44-r0 \
        libssl1.1=1.1.1l-r0 \
        libstdc++=10.3.1_git20210424-r2 \
        linux-headers=5.10.41-r0 \
        make=4.3-r0 \
        mpc1=1.2.1-r0 \
        mpfr4=4.1.0-r0 \
        musl=1.2.2-r3 \
        musl-dev=1.2.2-r3 \
        nghttp2-libs=1.43.0-r0 \
        openssl-dev=1.1.1l-r0 \
        openssl-libs-static=1.1.1l-r0 \
        patch=2.7.6-r7 \
        pcre=8.44-r0 \
        pcre-dev=8.44-r0 \
        pcre2=10.36-r0 \
        pkgconf=1.7.4-r0 \
        tzdata=2021d-r0 \
        zlib=1.2.11-r3 \
        zlib-dev=1.2.11-r3 \
        zlib-static=1.2.11-r3 \
        ; \
     git clone --depth=1 -b $TAG --single-branch git://github.com/nginx/nginx /src/nginx; \
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
      ln -s /dev/stderr /usr/local/nginx/logs/error.log

FROM reg.g5d.dev/base@sha256:bc61849aeec7da9f45293bb26ba1e37240ccb66a96c173ec8cfc817e71e4ce90

COPY --from=build --chown=65532:65532 /usr/local/nginx/logs /usr/local/nginx/logs
COPY --from=build --chown=65532:65532 /usr/local/nginx/scgi_temp /usr/local/nginx/scgi_temp
COPY --from=build --chown=65532:65532 /usr/local/nginx/uwsgi_temp /usr/local/nginx/uwsgi_tmp
COPY --from=build --chown=65532:65532 /usr/local/nginx/fastcgi_temp /usr/local/nginx/fastcgi_temp
COPY --from=build --chown=65532:65532 /usr/local/nginx/proxy_temp /usr/local/nginx/proxy_temp
COPY --from=build --chown=65532:65532 /usr/local/nginx/client_body_temp /usr/local/nginx/client_body_temp
COPY --from=build --chown=65532:65532 /usr/local/nginx/conf/mime.types /usr/local/nginx/conf/mime.types
COPY --from=build --chown=65532:65532 /usr/local/nginx/sbin/nginx /nginx

ENTRYPOINT ["/nginx"]
