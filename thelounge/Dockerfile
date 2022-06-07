# Generated from https://github.com/greboid/dockerfiles/blob/master/thelounge/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r6","apk:busybox":"1.35.0-r13","apk:ca-certificates":"20211220-r0","apk:expat":"2.4.8-r0","apk:git":"2.36.1-r0","apk:libcrypto1.1":"1.1.1o-r0","apk:libcurl":"7.83.1-r1","apk:libssl1.1":"1.1.1o-r0","apk:musl":"1.2.3-r0","apk:nghttp2-libs":"1.47.0-r0","apk:pcre2":"10.39-r0","apk:zlib":"1.2.12-r1","github:thelounge/thelounge":"v4.3.1","image:node":"2e9baff3a032a1a77742344d3172bcea07521a884c1296502e3cc60175216087"}

FROM reg.g5d.dev/node@sha256:2e9baff3a032a1a77742344d3172bcea07521a884c1296502e3cc60175216087 as yarnbase

WORKDIR /app

RUN corepack enable

FROM yarnbase as builder

ARG TAG="v4.3.1"

WORKDIR  /app
RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r6\
        busybox=1.35.0-r13\
        ca-certificates=20211220-r0\
        expat=2.4.8-r0\
        git=2.36.1-r0\
        libcrypto1.1=1.1.1o-r0\
        libcurl=7.83.1-r1\
        libssl1.1=1.1.1o-r0\
        musl=1.2.3-r0\
        nghttp2-libs=1.47.0-r0\
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
