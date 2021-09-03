# Generated from https://github.com/csmith/dockerfiles/blob/master/irc-news/Dockerfile.gotpl
# BOM: {"github:csmith/ircplugins":"v1.0.0","image:base":"sha256:1cc5b31fea33445d7b10537f4e5c2682fdc72e806b7c92dbdd3930a88fed86f0","image:golang":"sha256:b49ff0d1a458cc16f0a773ffffa321568e689f2abd2250a181b95c0cfb0c3567"}

FROM reg.g5d.dev/golang@sha256:b49ff0d1a458cc16f0a773ffffa321568e689f2abd2250a181b95c0cfb0c3567 AS build

ARG TAG="v1.0.0"

RUN set -eux; \
    apk add --no-cache git; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/csmith/ircplugins /go/src/github.com/csmith/ircplugins; \
    cd /go/src/github.com/csmith/ircplugins; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main ./cmd/news; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/csmith/ircplugins/main;

FROM reg.g5d.dev/base@sha256:1cc5b31fea33445d7b10537f4e5c2682fdc72e806b7c92dbdd3930a88fed86f0

COPY --from=build /go/src/github.com/csmith/ircplugins/main /irc-news
ENTRYPOINT ["/irc-news"]
