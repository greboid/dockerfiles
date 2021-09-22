# Generated from https://github.com/csmith/dockerfiles/blob/master/renovate/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.33.1-r3","apk:c-ares":"1.17.2-r0","apk:ca-certificates":"20191127-r5","apk:expat":"2.4.1-r0","apk:git":"2.32.0-r0","apk:libcrypto1.1":"1.1.1l-r0","apk:libcurl":"7.79.0-r0","apk:libgcc":"10.3.1_git20210424-r2","apk:libssl1.1":"1.1.1l-r0","apk:libstdc++":"10.3.1_git20210424-r2","apk:musl":"1.2.2-r3","apk:nghttp2-libs":"1.43.0-r0","apk:nodejs":"14.17.6-r0","apk:pcre2":"10.36-r0","apk:yarn":"1.22.10-r0","apk:zlib":"1.2.11-r3","image:alpine":"sha256:7cdc7a220adf9635167126cfe627b6ff000c12ddc5a44c7c5796161a81dae302","image:golang":"sha256:092432223820de2928ae7bd05e069c52b69349e29250695f56b793e17c32d0a2"}

FROM reg.g5d.dev/alpine@sha256:7cdc7a220adf9635167126cfe627b6ff000c12ddc5a44c7c5796161a81dae302 as yarnbase

RUN set -eux; \
    apk add --no-cache \
            brotli-libs=1.0.9-r5\
            busybox=1.33.1-r3\
            c-ares=1.17.2-r0\
            ca-certificates=20191127-r5\
            libcrypto1.1=1.1.1l-r0\
            libgcc=10.3.1_git20210424-r2\
            libssl1.1=1.1.1l-r0\
            libstdc++=10.3.1_git20210424-r2\
            musl=1.2.2-r3\
            nghttp2-libs=1.43.0-r0\
            nodejs=14.17.6-r0\
            yarn=1.22.10-r0\
            zlib=1.2.11-r3\
            ;

FROM yarnbase as deps

ARG TAG="27.15.0"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r5\
        busybox=1.33.1-r3\
        ca-certificates=20191127-r5\
        expat=2.4.1-r0\
        git=2.32.0-r0\
        libcrypto1.1=1.1.1l-r0\
        libcurl=7.79.0-r0\
        libssl1.1=1.1.1l-r0\
        musl=1.2.2-r3\
        nghttp2-libs=1.43.0-r0\
        pcre2=10.36-r0\
        zlib=1.2.11-r3\
        ; \
        git clone --depth=1 -b $TAG --single-branch https://github.com/renovatebot/renovate /js/src/github.com/renovatebot/renovate; \
        cd /js/src/github.com/renovatebot/renovate; \
        yarn install; \
        yarn cache clean; \
        yarn build; \
        chmod +x dist/renovate.js; \
        ln -sf dist/renovate.js renovate; \
        touch --date=@0 /js/src/github.com/renovatebot/renovate

FROM yarnbase

ENV NODE_ENV=production;\
    PATH=/usr/local/go/bin:$PATH

WORKDIR /app
COPY --from=reg.g5d.dev/golang@sha256:092432223820de2928ae7bd05e069c52b69349e29250695f56b793e17c32d0a2 /usr/local/go /usr/local/go
COPY --from=deps /js/src/github.com/renovatebot/renovate /app
CMD ["/app/renovate"]
