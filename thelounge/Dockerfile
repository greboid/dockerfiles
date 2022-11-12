# Generated from https://github.com/greboid/dockerfiles/blob/master/thelounge/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r6","apk:busybox":"1.35.0-r17","apk:ca-certificates":"20220614-r0","apk:expat":"2.5.0-r0","apk:git":"2.36.3-r0","apk:libcrypto1.1":"1.1.1s-r0","apk:libcurl":"7.83.1-r4","apk:libssl1.1":"1.1.1s-r0","apk:musl":"1.2.3-r2","apk:nghttp2-libs":"1.47.0-r0","apk:pcre2":"10.40-r0","apk:zlib":"1.2.12-r3","github:thelounge/thelounge":"v4.3.1","image:node":"be5dc75f07130656e92b7fc8f82f6957a8e51f424e70fae6d0f0b6e0b682830e"}

FROM reg.g5d.dev/node@sha256:be5dc75f07130656e92b7fc8f82f6957a8e51f424e70fae6d0f0b6e0b682830e as yarnbase

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
        expat=2.5.0-r0\
        git=2.36.3-r0\
        libcrypto1.1=1.1.1s-r0\
        libcurl=7.83.1-r4\
        libssl1.1=1.1.1s-r0\
        musl=1.2.3-r2\
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
