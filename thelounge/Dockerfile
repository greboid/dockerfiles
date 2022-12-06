# Generated from https://github.com/greboid/dockerfiles/blob/master/thelounge/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r2","apk:git":"2.38.1-r0","apk:libcrypto3":"3.0.7-r0","apk:libcurl":"7.86.0-r1","apk:libexpat":"2.5.0-r0","apk:libssl3":"3.0.7-r0","apk:musl":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:pcre2":"10.40-r0","apk:zlib":"1.2.13-r0","github:thelounge/thelounge":"v4.3.1","image:node":"7f24f6e76ca24ae97c278990de51b0b1c3dd9ab8fdf8d25a0f9148695eb010d8"}

FROM reg.g5d.dev/node@sha256:7f24f6e76ca24ae97c278990de51b0b1c3dd9ab8fdf8d25a0f9148695eb010d8 as yarnbase

WORKDIR /app

RUN corepack enable

FROM yarnbase as builder

ARG TAG="v4.3.1"

WORKDIR  /app
RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r9\
        busybox=1.35.0-r29\
        busybox-binsh=1.35.0-r29\
        ca-certificates=20220614-r2\
        git=2.38.1-r0\
        libcrypto3=3.0.7-r0\
        libcurl=7.86.0-r1\
        libexpat=2.5.0-r0\
        libssl3=3.0.7-r0\
        musl=1.2.3-r4\
        nghttp2-libs=1.51.0-r0\
        pcre2=10.40-r0\
        zlib=1.2.13-r0\
        ; \
    yarn global add thelounge@$TAG; \
    mkdir /home/nonroot; \
    chown -R 65532:65532 /home/nonroot

EXPOSE 9000
ENV THELOUNGE_HOME /home/nonroot
WORKDIR /home/nonroot

USER 65532

ENTRYPOINT ["/usr/local/bin/thelounge"]
