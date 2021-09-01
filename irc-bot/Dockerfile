# Generated from https://github.com/csmith/dockerfiles/blob/master/irc-bot/Dockerfile.gotpl
# BOM: {"github:greboid/irc-bot":"v5.0.1","image:base":"sha256:1cc5b31fea33445d7b10537f4e5c2682fdc72e806b7c92dbdd3930a88fed86f0","image:golang":"sha256:dd04c1c827e6f4db0d32bbd80eac84f8899bb252dd1961f64fdacd279d369a71"}

FROM reg.g5d.dev/golang@sha256:dd04c1c827e6f4db0d32bbd80eac84f8899bb252dd1961f64fdacd279d369a71 AS build

ARG TAG="v5.0.1"

RUN set -eux; \
    apk add --no-cache git; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/greboid/irc-bot /go/src/github.com/greboid/irc-bot; \
    cd /go/src/github.com/greboid/irc-bot; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main ./cmd/bot; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/greboid/irc-bot/main;

FROM reg.g5d.dev/base@sha256:1cc5b31fea33445d7b10537f4e5c2682fdc72e806b7c92dbdd3930a88fed86f0

COPY --from=build /go/src/github.com/greboid/irc-bot/main /irc-bot
EXPOSE 8080
ENTRYPOINT ["/irc-bot"]
