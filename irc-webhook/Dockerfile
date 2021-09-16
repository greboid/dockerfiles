# Generated from https://github.com/csmith/dockerfiles/blob/master/irc-webhook/Dockerfile.gotpl
# BOM: {"github:greboid/irc-webhook":"v4.0.1","image:base":"sha256:9eb1693b711c4282b96c772aecfb3fe764a384ed907fb24f80831830e90678c8","image:golang":"sha256:c1d9d794a310682b45fbdaad1dbe94ee3773ce218c3c0fcdff6b499cc87c6f4c"}

FROM reg.g5d.dev/golang@sha256:c1d9d794a310682b45fbdaad1dbe94ee3773ce218c3c0fcdff6b499cc87c6f4c AS build

ARG TAG="v4.0.1"

RUN set -eux; \
    apk add --no-cache git; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/greboid/irc-webhook /go/src/github.com/greboid/irc-webhook; \
    cd /go/src/github.com/greboid/irc-webhook; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main ./cmd/webhook; \
    go run github.com/google/go-licenses@latest save ./... --save_path=/notices; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/greboid/irc-webhook/main;

FROM reg.g5d.dev/base@sha256:9eb1693b711c4282b96c772aecfb3fe764a384ed907fb24f80831830e90678c8

COPY --from=build /go/src/github.com/greboid/irc-webhook/main /irc-webhook
COPY --from=build /notices /notices
ENTRYPOINT ["/irc-webhook"]
