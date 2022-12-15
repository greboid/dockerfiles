# Generated from https://github.com/greboid/dockerfiles/blob/master/renovate/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:c-ares":"1.18.1-r1","apk:ca-certificates":"20220614-r2","apk:git":"2.38.2-r0","apk:icu-data-en":"72.1-r1","apk:icu-libs":"72.1-r1","apk:libcrypto3":"3.0.7-r0","apk:libcurl":"7.86.0-r1","apk:libexpat":"2.5.0-r0","apk:libgcc":"12.2.1_git20220924-r4","apk:libssl3":"3.0.7-r0","apk:libstdc++":"12.2.1_git20220924-r4","apk:musl":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:nodejs":"18.12.1-r0","apk:pcre2":"10.42-r0","apk:yarn":"1.22.19-r0","apk:zlib":"1.2.13-r0","image:alpine":"a511477ce0dad879d21b044feb8bcb146eb420de4982cd2b9aafa1347ee7a2ff","image:golang":"4b4ba7b1f9b635c33fe5c6abc66075d02ee7bd1e67d62421169dbcd76d8d81bf"}

FROM reg.g5d.dev/alpine@sha256:a511477ce0dad879d21b044feb8bcb146eb420de4982cd2b9aafa1347ee7a2ff as yarnbase

RUN set -eux; \
    apk add --no-cache \
            brotli-libs=1.0.9-r9\
            busybox=1.35.0-r29\
            busybox-binsh=1.35.0-r29\
            c-ares=1.18.1-r1\
            ca-certificates=20220614-r2\
            git=2.38.2-r0\
            icu-data-en=72.1-r1\
            icu-libs=72.1-r1\
            libcrypto3=3.0.7-r0\
            libcurl=7.86.0-r1\
            libexpat=2.5.0-r0\
            libgcc=12.2.1_git20220924-r4\
            libssl3=3.0.7-r0\
            libstdc++=12.2.1_git20220924-r4\
            musl=1.2.3-r4\
            nghttp2-libs=1.51.0-r0\
            nodejs=18.12.1-r0\
            pcre2=10.42-r0\
            yarn=1.22.19-r0\
            zlib=1.2.13-r0\
            ;

FROM yarnbase as deps

ARG TAG="27.15.0"

RUN set -eux; \
        git clone --depth=1 -b $TAG --single-branch https://github.com/renovatebot/renovate /js/src/github.com/renovatebot/renovate; \
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
COPY --from=reg.g5d.dev/golang@sha256:4b4ba7b1f9b635c33fe5c6abc66075d02ee7bd1e67d62421169dbcd76d8d81bf /usr/local/go /usr/local/go
COPY --from=deps /js/src/github.com/renovatebot/renovate /app
CMD ["/app/renovate"]
