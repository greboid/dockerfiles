# Generated from https://github.com/greboid/dockerfiles/blob/master/renovate/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r6","apk:busybox":"1.35.0-r14","apk:c-ares":"1.18.1-r0","apk:ca-certificates":"20211220-r0","apk:expat":"2.4.8-r0","apk:git":"2.36.1-r0","apk:icu-data-en":"71.1-r2","apk:icu-libs":"71.1-r2","apk:libcrypto1.1":"1.1.1q-r0","apk:libcurl":"7.83.1-r2","apk:libgcc":"11.2.1_git20220219-r2","apk:libssl1.1":"1.1.1q-r0","apk:libstdc++":"11.2.1_git20220219-r2","apk:musl":"1.2.3-r0","apk:nghttp2-libs":"1.47.0-r0","apk:nodejs":"16.15.0-r1","apk:pcre2":"10.40-r0","apk:yarn":"1.22.19-r0","apk:zlib":"1.2.12-r1","image:alpine":"e1af61331d46ce7a5770c5083e1c81e297fbf3fcc97b63542b0ac62a8f3900a3","image:golang":"acbd92535de0eedba3a1d5469ed30ba66cc7b6b6d105845498df382ff1db5a93"}

FROM reg.g5d.dev/alpine@sha256:e1af61331d46ce7a5770c5083e1c81e297fbf3fcc97b63542b0ac62a8f3900a3 as yarnbase

RUN set -eux; \
    apk add --no-cache \
            brotli-libs=1.0.9-r6\
            busybox=1.35.0-r14\
            c-ares=1.18.1-r0\
            ca-certificates=20211220-r0\
            expat=2.4.8-r0\
            git=2.36.1-r0\
            icu-data-en=71.1-r2\
            icu-libs=71.1-r2\
            libcrypto1.1=1.1.1q-r0\
            libcurl=7.83.1-r2\
            libgcc=11.2.1_git20220219-r2\
            libssl1.1=1.1.1q-r0\
            libstdc++=11.2.1_git20220219-r2\
            musl=1.2.3-r0\
            nghttp2-libs=1.47.0-r0\
            nodejs=16.15.0-r1\
            pcre2=10.40-r0\
            yarn=1.22.19-r0\
            zlib=1.2.12-r1\
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
COPY --from=reg.g5d.dev/golang@sha256:acbd92535de0eedba3a1d5469ed30ba66cc7b6b6d105845498df382ff1db5a93 /usr/local/go /usr/local/go
COPY --from=deps /js/src/github.com/renovatebot/renovate /app
CMD ["/app/renovate"]
