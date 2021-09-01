# Generated from https://github.com/csmith/dockerfiles/blob/master/irc-distribution/Dockerfile.gotpl
# BOM: {"github:csmith/irc-distribution":"v1.0.0","image:base":"sha256:1cc5b31fea33445d7b10537f4e5c2682fdc72e806b7c92dbdd3930a88fed86f0","image:golang":"sha256:dd04c1c827e6f4db0d32bbd80eac84f8899bb252dd1961f64fdacd279d369a71"}

FROM reg.g5d.dev/golang@sha256:dd04c1c827e6f4db0d32bbd80eac84f8899bb252dd1961f64fdacd279d369a71 AS build

ARG TAG="v1.0.0"

RUN set -eux; \
    apk add --no-cache git; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/csmith/irc-distribution /go/src/github.com/csmith/irc-distribution; \
    cd /go/src/github.com/csmith/irc-distribution; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main ./cmd/distribution; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/csmith/irc-distribution/main;

FROM reg.g5d.dev/base@sha256:1cc5b31fea33445d7b10537f4e5c2682fdc72e806b7c92dbdd3930a88fed86f0

COPY --from=build /go/src/github.com/csmith/irc-distribution/main /irc-distribution
ENTRYPOINT ["/irc-distribution"]
