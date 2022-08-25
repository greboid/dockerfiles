# Generated from https://github.com/greboid/dockerfiles/blob/master/thelounge/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r6","apk:busybox":"1.35.0-r17","apk:ca-certificates":"20220614-r0","apk:expat":"2.4.8-r0","apk:git":"2.36.2-r0","apk:libcrypto1.1":"1.1.1q-r0","apk:libcurl":"7.83.1-r2","apk:libssl1.1":"1.1.1q-r0","apk:musl":"1.2.3-r0","apk:nghttp2-libs":"1.47.0-r0","apk:pcre2":"10.40-r0","apk:zlib":"1.2.12-r3","github:thelounge/thelounge":"v4.3.1","image:node":"e775e47831a954d1a16dfe1ad23392983c02453a10c09bdd651ea4c8959770cb"}

FROM reg.g5d.dev/node@sha256:e775e47831a954d1a16dfe1ad23392983c02453a10c09bdd651ea4c8959770cb as yarnbase

WORKDIR /app

RUN corepack enable

FROM yarnbase as builder

ARG TAG="v4.3.1"

WORKDIR  /app
RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r6\
        busybox=1.35.0-r17\
        ca-certificates=20220614-r0\
        expat=2.4.8-r0\
        git=2.36.2-r0\
        libcrypto1.1=1.1.1q-r0\
        libcurl=7.83.1-r2\
        libssl1.1=1.1.1q-r0\
        musl=1.2.3-r0\
        nghttp2-libs=1.47.0-r0\
        pcre2=10.40-r0\
        zlib=1.2.12-r3\
        ; \
    yarn global add thelounge@$TAG; \
    mkdir /home/nonroot; \
    chown -R 65532:65532 /home/nonroot

EXPOSE 9000
ENV THELOUNGE_HOME /home/nonroot
WORKDIR /home/nonroot

USER 65532

ENTRYPOINT ["/usr/local/bin/thelounge"]
