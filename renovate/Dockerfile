# Generated from https://github.com/greboid/dockerfiles/blob/master/renovate/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.34.1-r5","apk:c-ares":"1.18.1-r0","apk:ca-certificates":"20211220-r0","apk:expat":"2.4.7-r0","apk:git":"2.34.2-r0","apk:libcrypto1.1":"1.1.1n-r0","apk:libcurl":"7.80.0-r0","apk:libgcc":"10.3.1_git20211027-r0","apk:libssl1.1":"1.1.1n-r0","apk:libstdc++":"10.3.1_git20211027-r0","apk:musl":"1.2.2-r7","apk:nghttp2-libs":"1.46.0-r0","apk:nodejs":"16.14.2-r0","apk:pcre2":"10.39-r0","apk:yarn":"1.22.17-r0","apk:zlib":"1.2.12-r0","image:alpine":"9f06ed7307f2d03a3d486b9b0bfe7737dfe3c7b3e2f8e906f5a2fcc9408867ac","image:golang":"45063dd5cfeb92e1954b68c879174f0ddad072c00a6ebac396e7685f312e4a19"}

FROM reg.g5d.dev/alpine@sha256:9f06ed7307f2d03a3d486b9b0bfe7737dfe3c7b3e2f8e906f5a2fcc9408867ac as yarnbase

RUN set -eux; \
    apk add --no-cache \
            brotli-libs=1.0.9-r5\
            busybox=1.34.1-r5\
            c-ares=1.18.1-r0\
            ca-certificates=20211220-r0\
            expat=2.4.7-r0\
            git=2.34.2-r0\
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
            zlib=1.2.12-r0\
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
COPY --from=reg.g5d.dev/golang@sha256:45063dd5cfeb92e1954b68c879174f0ddad072c00a6ebac396e7685f312e4a19 /usr/local/go /usr/local/go
COPY --from=deps /js/src/github.com/renovatebot/renovate /app
CMD ["/app/renovate"]
