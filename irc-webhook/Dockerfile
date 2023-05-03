# Generated from https://github.com/greboid/dockerfiles/blob/master/irc-webhook/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r4","apk:git":"2.38.5-r0","apk:libcrypto3":"3.0.8-r4","apk:libcurl":"8.0.1-r0","apk:libexpat":"2.5.0-r0","apk:libssl3":"3.0.8-r4","apk:musl":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:pcre2":"10.42-r0","apk:zlib":"1.2.13-r0","github:greboid/irc-webhook":"v4.0.1","image:base":"ef8b65fda5d776551a87f970b02ebd1aa67f98c60b68637d910429a4e18be9b9","image:golang":"b792ccdfd92b7088b37da03e87180e53a5835ac1cf0fa128cce9ed0f2799c088"}

FROM reg.g5d.dev/golang@sha256:b792ccdfd92b7088b37da03e87180e53a5835ac1cf0fa128cce9ed0f2799c088 AS build

ARG TAG="v4.0.1"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r9\
        busybox=1.35.0-r29\
        busybox-binsh=1.35.0-r29\
        ca-certificates=20220614-r4\
        git=2.38.5-r0\
        libcrypto3=3.0.8-r4\
        libcurl=8.0.1-r0\
        libexpat=2.5.0-r0\
        libssl3=3.0.8-r4\
        musl=1.2.3-r4\
        nghttp2-libs=1.51.0-r0\
        pcre2=10.42-r0\
        zlib=1.2.13-r0\
        ; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/greboid/irc-webhook /go/src/github.com/greboid/irc-webhook; \
    cd /go/src/github.com/greboid/irc-webhook; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main ./cmd/webhook; \
    go-licenses save ./... --save_path=/notices;

FROM reg.g5d.dev/base@sha256:ef8b65fda5d776551a87f970b02ebd1aa67f98c60b68637d910429a4e18be9b9

COPY --from=build /go/src/github.com/greboid/irc-webhook/main /irc-webhook
COPY --from=build /notices /notices
ENTRYPOINT ["/irc-webhook"]
