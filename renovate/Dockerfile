# Generated from https://github.com/greboid/dockerfiles/blob/master/renovate/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r6","apk:busybox":"1.35.0-r17","apk:c-ares":"1.18.1-r0","apk:ca-certificates":"20220614-r0","apk:expat":"2.4.9-r0","apk:git":"2.36.2-r0","apk:icu-data-en":"71.1-r2","apk:icu-libs":"71.1-r2","apk:libcrypto1.1":"1.1.1q-r0","apk:libcurl":"7.83.1-r3","apk:libgcc":"11.2.1_git20220219-r2","apk:libssl1.1":"1.1.1q-r0","apk:libstdc++":"11.2.1_git20220219-r2","apk:musl":"1.2.3-r0","apk:nghttp2-libs":"1.47.0-r0","apk:nodejs":"16.16.0-r0","apk:pcre2":"10.40-r0","apk:yarn":"1.22.19-r0","apk:zlib":"1.2.12-r3","image:alpine":"a475fe9f28063dd6914342b78877af564ac1144cacba56ddb2aed1bce0c3fc14","image:golang":"452076da215234bbba5bc52c7e8b96e3dcaea369f630de58a6e7e9ca33875290"}

FROM reg.g5d.dev/alpine@sha256:a475fe9f28063dd6914342b78877af564ac1144cacba56ddb2aed1bce0c3fc14 as yarnbase

RUN set -eux; \
    apk add --no-cache \
            brotli-libs=1.0.9-r6\
            busybox=1.35.0-r17\
            c-ares=1.18.1-r0\
            ca-certificates=20220614-r0\
            expat=2.4.9-r0\
            git=2.36.2-r0\
            icu-data-en=71.1-r2\
            icu-libs=71.1-r2\
            libcrypto1.1=1.1.1q-r0\
            libcurl=7.83.1-r3\
            libgcc=11.2.1_git20220219-r2\
            libssl1.1=1.1.1q-r0\
            libstdc++=11.2.1_git20220219-r2\
            musl=1.2.3-r0\
            nghttp2-libs=1.47.0-r0\
            nodejs=16.16.0-r0\
            pcre2=10.40-r0\
            yarn=1.22.19-r0\
            zlib=1.2.12-r3\
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
COPY --from=reg.g5d.dev/golang@sha256:452076da215234bbba5bc52c7e8b96e3dcaea369f630de58a6e7e9ca33875290 /usr/local/go /usr/local/go
COPY --from=deps /js/src/github.com/renovatebot/renovate /app
CMD ["/app/renovate"]
