# Generated from https://github.com/csmith/dockerfiles/blob/master/irc-news/Dockerfile.gotpl
# BOM: {"github:csmith/ircplugins":"v1.0.0","image:base":"sha256:f7f27db7afb58bae23ad902072228f1b090b878af303d9f63bd2c1526b9b4f53","image:golang":"sha256:beae544bfe746af2dec32e19106ea3ce7e346dde3c9692bb1a82c70be9b636e4"}

FROM reg.g5d.dev/golang@sha256:beae544bfe746af2dec32e19106ea3ce7e346dde3c9692bb1a82c70be9b636e4 AS build

ARG TAG="v1.0.0"

RUN set -eux; \
    apk add --no-cache git; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/csmith/ircplugins /go/src/github.com/csmith/ircplugins; \
    cd /go/src/github.com/csmith/ircplugins; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main ./cmd/news; \
    go run github.com/google/go-licenses@latest save ./... --save_path=/notices; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/csmith/ircplugins/main;

FROM reg.g5d.dev/base@sha256:f7f27db7afb58bae23ad902072228f1b090b878af303d9f63bd2c1526b9b4f53

COPY --from=build /go/src/github.com/csmith/ircplugins/main /irc-news
COPY --from=build /notices /notices
ENTRYPOINT ["/irc-news"]
