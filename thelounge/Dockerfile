# Generated from https://github.com/greboid/dockerfiles/blob/master/thelounge/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r14","apk:busybox":"1.36.0-r9","apk:busybox-binsh":"1.36.0-r9","apk:ca-certificates":"20230506-r0","apk:git":"2.40.1-r0","apk:libcrypto3":"3.1.1-r1","apk:libcurl":"8.1.2-r0","apk:libexpat":"2.5.0-r1","apk:libidn2":"2.3.4-r1","apk:libssl3":"3.1.1-r1","apk:libunistring":"1.1-r1","apk:musl":"1.2.4-r0","apk:nghttp2-libs":"1.53.0-r0","apk:pcre2":"10.42-r1","apk:zlib":"1.2.13-r1","github:thelounge/thelounge":"v4.4.0","image:node":"5e54bd343643bb1c491648791c46813f5336d3f0339bd928ef1d717635bf7559"}

FROM reg.g5d.dev/node@sha256:5e54bd343643bb1c491648791c46813f5336d3f0339bd928ef1d717635bf7559 as yarnbase

WORKDIR /app

RUN corepack enable

FROM yarnbase as builder

ARG TAG="v4.4.0"

WORKDIR  /app
RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r14\
        busybox=1.36.0-r9\
        busybox-binsh=1.36.0-r9\
        ca-certificates=20230506-r0\
        git=2.40.1-r0\
        libcrypto3=3.1.1-r1\
        libcurl=8.1.2-r0\
        libexpat=2.5.0-r1\
        libidn2=2.3.4-r1\
        libssl3=3.1.1-r1\
        libunistring=1.1-r1\
        musl=1.2.4-r0\
        nghttp2-libs=1.53.0-r0\
        pcre2=10.42-r1\
        zlib=1.2.13-r1\
        ; \
    yarn global add thelounge@$TAG; \
    mkdir /home/nonroot; \
    chown -R 65532:65532 /home/nonroot

EXPOSE 9000
ENV THELOUNGE_HOME /home/nonroot
WORKDIR /home/nonroot

USER 65532

ENTRYPOINT ["/usr/local/bin/thelounge"]
