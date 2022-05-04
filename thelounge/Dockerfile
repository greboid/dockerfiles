# Generated from https://github.com/greboid/dockerfiles/blob/master/thelounge/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.34.1-r5","apk:ca-certificates":"20211220-r0","apk:expat":"2.4.7-r0","apk:git":"2.34.2-r0","apk:libcrypto1.1":"1.1.1n-r0","apk:libcurl":"7.80.0-r1","apk:libssl1.1":"1.1.1n-r0","apk:musl":"1.2.2-r7","apk:nghttp2-libs":"1.46.0-r0","apk:pcre2":"10.39-r0","apk:zlib":"1.2.12-r1","github:thelounge/thelounge":"v4.3.1","image:node":"5a3e682d2e862a4f87cb3821ec6efc70f961f324b6931697a73361c464261215"}

FROM reg.g5d.dev/node@sha256:5a3e682d2e862a4f87cb3821ec6efc70f961f324b6931697a73361c464261215 as yarnbase

WORKDIR /app

RUN corepack enable

FROM yarnbase as builder

ARG TAG="v4.3.1"

WORKDIR  /app
RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r5\
        busybox=1.34.1-r5\
        ca-certificates=20211220-r0\
        expat=2.4.7-r0\
        git=2.34.2-r0\
        libcrypto1.1=1.1.1n-r0\
        libcurl=7.80.0-r1\
        libssl1.1=1.1.1n-r0\
        musl=1.2.2-r7\
        nghttp2-libs=1.46.0-r0\
        pcre2=10.39-r0\
        zlib=1.2.12-r1\
        ; \
    yarn global add thelounge@4.3.0; \
    mkdir /home/nonroot; \
    chown -R 65532:65532 /home/nonroot

EXPOSE 9000
ENV THELOUNGE_HOME /home/nonroot
WORKDIR /home/nonroot

USER 65532

ENTRYPOINT ["/usr/local/bin/thelounge"]
