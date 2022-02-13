# Generated from https://github.com/greboid/dockerfiles/blob/master/greboid.com/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.34.1-r4","apk:ca-certificates":"20211220-r0","apk:expat":"2.4.4-r0","apk:giflib":"5.2.1-r0","apk:git":"2.34.1-r0","apk:libcrypto1.1":"1.1.1l-r8","apk:libcurl":"7.80.0-r0","apk:libjpeg-turbo":"2.1.2-r0","apk:libpng":"1.6.37-r1","apk:libssl1.1":"1.1.1l-r8","apk:libwebp":"1.2.2-r0","apk:libwebp-tools":"1.2.2-r0","apk:musl":"1.2.2-r7","apk:nghttp2-libs":"1.46.0-r0","apk:pcre2":"10.39-r0","apk:zlib":"1.2.11-r3","github:greboid/greboid.com":"v1.0.4","image:alpine":"ef95559f475cdf56e55db71c168f1616d497e2215bbd9f6bc63398036ae546a1","image:ghcr.io/greboid/cv":"7b4ca0e659e62a03d8860bbf553d923ae44348fe3f0c8327e8c38f3f0aa6dcc2","image:hugo":"f36cb2b43d70938735b1e069d222e4d258bd2dd78b43398d787fc8e7d859523f","image:nginx":"5c0e344dd3c5e4e53d7c4cbf3194416e814c1c56a15f04221147e67cc2b0b6e7"}

FROM ghcr.io/greboid/cv@sha256:7b4ca0e659e62a03d8860bbf553d923ae44348fe3f0c8327e8c38f3f0aa6dcc2 as cv

FROM reg.g5d.dev/alpine@sha256:ef95559f475cdf56e55db71c168f1616d497e2215bbd9f6bc63398036ae546a1 as site

ARG TAG="v1.0.4"

RUN set -eux; \
    apk add --no-cache --virtual .build-deps \
    brotli-libs=1.0.9-r5 \
    busybox=1.34.1-r4 \
    ca-certificates=20211220-r0 \
    expat=2.4.4-r0 \
    git=2.34.1-r0 \
    libcrypto1.1=1.1.1l-r8 \
    libcurl=7.80.0-r0 \
    libssl1.1=1.1.1l-r8 \
    musl=1.2.2-r7 \
    nghttp2-libs=1.46.0-r0 \
    pcre2=10.39-r0 \
    zlib=1.2.11-r3 \
    ; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/greboid/greboid.com /build;

FROM reg.g5d.dev/hugo@sha256:f36cb2b43d70938735b1e069d222e4d258bd2dd78b43398d787fc8e7d859523f as hugo

COPY --from=site --chown=65532:65532 /build /build
COPY --from=cv --chown=65532:65532 /cv.pdf /build/site/static/cv.pdf
RUN ["/usr/bin/hugo", "-s", "/build/site", "-d", "/build/public"]

FROM reg.g5d.dev/alpine@sha256:ef95559f475cdf56e55db71c168f1616d497e2215bbd9f6bc63398036ae546a1 as minify
RUN set -eux; \
    apk add --no-cache --virtual .build-deps \
    giflib=5.2.1-r0 \
    libjpeg-turbo=2.1.2-r0 \
    libpng=1.6.37-r1 \
    libwebp=1.2.2-r0 \
    libwebp-tools=1.2.2-r0 \
    musl=1.2.2-r7 \
    zlib=1.2.11-r3 \
    ;

COPY --from=hugo --chown=65532:65532 /build/public /tmp/public
USER 65532:65532
RUN set -eux; \
    find /tmp/public \( -name '*.jpg' -o -name '*.png' -o -name '*.jpeg' \) -exec cwebp -q 60 "{}" -o "{}.webp" \;;

FROM reg.g5d.dev/nginx@sha256:5c0e344dd3c5e4e53d7c4cbf3194416e814c1c56a15f04221147e67cc2b0b6e7 AS nginx
COPY --from=minify /tmp/public /src/public
COPY --from=site /build/nginx.conf /usr/local/nginx/conf/nginx.conf

ENTRYPOINT ["/nginx"]
