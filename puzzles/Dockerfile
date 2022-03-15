# Generated from https://github.com/greboid/dockerfiles/blob/master/puzzles/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.34.1-r4","apk:ca-certificates":"20211220-r0","apk:expat":"2.4.7-r0","apk:git":"2.34.1-r0","apk:libcrypto1.1":"1.1.1l-r8","apk:libcurl":"7.80.0-r0","apk:libssl1.1":"1.1.1l-r8","apk:musl":"1.2.2-r7","apk:nghttp2-libs":"1.46.0-r0","apk:pcre2":"10.39-r0","apk:zlib":"1.2.11-r3","github:greboid/puzzles":"v1.0.0","image:base":"e390636a8773e2b84066cdbde990e80b3c97b8e7e25428df9f4b8f097323f861","image:golang":"4d334f9039eeb657925628b9ebb364944bbe6848b8bf98b3b57091fc971fc451"}

FROM reg.g5d.dev/golang@sha256:4d334f9039eeb657925628b9ebb364944bbe6848b8bf98b3b57091fc971fc451 AS build

ARG TAG="v1.0.0"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r5\
        busybox=1.34.1-r4\
        ca-certificates=20211220-r0\
        expat=2.4.7-r0\
        git=2.34.1-r0\
        libcrypto1.1=1.1.1l-r8\
        libcurl=7.80.0-r0\
        libssl1.1=1.1.1l-r8\
        musl=1.2.2-r7\
        nghttp2-libs=1.46.0-r0\
        pcre2=10.39-r0\
        zlib=1.2.11-r3\
        ; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/greboid/puzzles /go/src/github.com/greboid/puzzles; \
    cd /go/src/github.com/greboid/puzzles; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    go run github.com/google/go-licenses@latest save ./... --save_path=/notices; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/greboid/puzzles/main;

FROM reg.g5d.dev/base@sha256:e390636a8773e2b84066cdbde990e80b3c97b8e7e25428df9f4b8f097323f861

COPY --from=build /go/src/github.com/greboid/puzzles/main /puzzles
COPY --from=build /go/src/github.com/greboid/puzzles/wordlists/. /app/wordlists
COPY --from=build /notices /notices
EXPOSE 8080
ENTRYPOINT ["/puzzles"]
