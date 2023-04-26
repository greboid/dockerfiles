# Generated from https://github.com/greboid/dockerfiles/blob/master/puzzles/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r4","apk:git":"2.38.5-r0","apk:libcrypto3":"3.0.8-r4","apk:libcurl":"8.0.1-r0","apk:libexpat":"2.5.0-r0","apk:libssl3":"3.0.8-r4","apk:musl":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:pcre2":"10.42-r0","apk:zlib":"1.2.13-r0","github:greboid/puzzles":"v1.0.1","image:base":"ef8b65fda5d776551a87f970b02ebd1aa67f98c60b68637d910429a4e18be9b9","image:golang":"98c2613d842792ed13dd0907562a814c4c47d03acb877560544f4b62988e7b82"}

FROM reg.g5d.dev/golang@sha256:98c2613d842792ed13dd0907562a814c4c47d03acb877560544f4b62988e7b82 AS build

ARG TAG="v1.0.1"

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
    git clone --depth=1 -b $TAG --single-branch https://github.com/greboid/puzzles /go/src/github.com/greboid/puzzles; \
    cd /go/src/github.com/greboid/puzzles; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    go-licenses save ./... --save_path=/notices;

FROM reg.g5d.dev/base@sha256:ef8b65fda5d776551a87f970b02ebd1aa67f98c60b68637d910429a4e18be9b9

COPY --from=build /go/src/github.com/greboid/puzzles/main /puzzles
COPY --from=build /go/src/github.com/greboid/puzzles/wordlists/. /app/wordlists
COPY --from=build /notices /notices
EXPOSE 8080
ENTRYPOINT ["/puzzles"]
