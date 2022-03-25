# Generated from https://github.com/greboid/dockerfiles/blob/master/renovate/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.34.1-r4","apk:c-ares":"1.18.1-r0","apk:ca-certificates":"20211220-r0","apk:expat":"2.4.7-r0","apk:git":"2.34.1-r0","apk:libcrypto1.1":"1.1.1n-r0","apk:libcurl":"7.80.0-r0","apk:libgcc":"10.3.1_git20211027-r0","apk:libssl1.1":"1.1.1n-r0","apk:libstdc++":"10.3.1_git20211027-r0","apk:musl":"1.2.2-r7","apk:nghttp2-libs":"1.46.0-r0","apk:nodejs":"16.14.2-r0","apk:pcre2":"10.39-r0","apk:yarn":"1.22.17-r0","apk:zlib":"1.2.11-r3","image:alpine":"a4460332f57a36a82ca308dd52b6565829a025b66580248508b45908f0061bc1","image:golang":"328e0e32945e05d5843e63fb286c7474c089614cc195bd0308255b61f3f8a1c2"}

FROM reg.g5d.dev/alpine@sha256:a4460332f57a36a82ca308dd52b6565829a025b66580248508b45908f0061bc1 as yarnbase

RUN set -eux; \
    apk add --no-cache \
            brotli-libs=1.0.9-r5\
            busybox=1.34.1-r4\
            c-ares=1.18.1-r0\
            ca-certificates=20211220-r0\
            expat=2.4.7-r0\
            git=2.34.1-r0\
            libcrypto1.1=1.1.1n-r0\
            libcurl=7.80.0-r0\
            libgcc=10.3.1_git20211027-r0\
            libssl1.1=1.1.1n-r0\
            libstdc++=10.3.1_git20211027-r0\
            musl=1.2.2-r7\
            nghttp2-libs=1.46.0-r0\
            nodejs=16.14.2-r0\
            pcre2=10.39-r0\
            yarn=1.22.17-r0\
            zlib=1.2.11-r3\
            ;

FROM yarnbase as deps

ARG TAG="27.15.0"

RUN set -eux; \
        git -c advice.detachedHead=false clone --depth=1 -b $TAG --single-branch https://github.com/renovatebot/renovate /js/src/github.com/renovatebot/renovate; \
        cd /js/src/github.com/renovatebot/renovate; \
        yarn install; \
        yarn cache clean; \
        yarn build; \
        chmod +x dist/renovate.js; \
        ln -sf dist/renovate.js renovate;

FROM yarnbase

ENV NODE_ENV=production;\
    PATH=/usr/local/go/bin:$PATH

WORKDIR /app
COPY --from=reg.g5d.dev/golang@sha256:328e0e32945e05d5843e63fb286c7474c089614cc195bd0308255b61f3f8a1c2 /usr/local/go /usr/local/go
COPY --from=deps /js/src/github.com/renovatebot/renovate /app
CMD ["/app/renovate"]
