# Generated from https://github.com/csmith/dockerfiles/blob/master/irc-bot/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.33.1-r3","apk:ca-certificates":"20191127-r5","apk:expat":"2.4.1-r0","apk:git":"2.32.0-r0","apk:libcrypto1.1":"1.1.1l-r0","apk:libcurl":"7.79.1-r0","apk:libssl1.1":"1.1.1l-r0","apk:musl":"1.2.2-r3","apk:nghttp2-libs":"1.43.0-r0","apk:pcre2":"10.36-r0","apk:zlib":"1.2.11-r3","github:greboid/irc-bot":"v5.0.4","image:base":"sha256:93cb0a11840ca0bf96975b0e9303a234bae7988481533300702d12a5857b630a","image:golang":"sha256:65e504b0cb4e5df85e2301f47cd3f231768d7b0d5aba59b1201e9c50fdf5e0ac"}

FROM reg.g5d.dev/golang@sha256:65e504b0cb4e5df85e2301f47cd3f231768d7b0d5aba59b1201e9c50fdf5e0ac AS build

ARG TAG="v5.0.4"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r5\
        busybox=1.33.1-r3\
        ca-certificates=20191127-r5\
        expat=2.4.1-r0\
        git=2.32.0-r0\
        libcrypto1.1=1.1.1l-r0\
        libcurl=7.79.1-r0\
        libssl1.1=1.1.1l-r0\
        musl=1.2.2-r3\
        nghttp2-libs=1.43.0-r0\
        pcre2=10.36-r0\
        zlib=1.2.11-r3\
        ; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/greboid/irc-bot /go/src/github.com/greboid/irc-bot; \
    cd /go/src/github.com/greboid/irc-bot; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main ./cmd/bot; \
    go run github.com/google/go-licenses@latest save ./... --save_path=/notices; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/greboid/irc-bot/main;

FROM reg.g5d.dev/base@sha256:93cb0a11840ca0bf96975b0e9303a234bae7988481533300702d12a5857b630a

COPY --from=build /go/src/github.com/greboid/irc-bot/main /irc-bot
COPY --from=build /notices /notices
EXPOSE 8080
ENTRYPOINT ["/irc-bot"]
