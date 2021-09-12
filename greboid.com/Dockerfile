# Generated from https://github.com/csmith/dockerfiles/blob/master/greboid.com/Dockerfile.gotpl
# BOM: {"github:greboid/greboid.com":"v1.0.1","image:base":"sha256:9eb1693b711c4282b96c772aecfb3fe764a384ed907fb24f80831830e90678c8","image:golang":"sha256:176928588492da996e1d64b66397ca6c3e7b629d2ba3e333f2b0ae10bffe637b"}

FROM reg.g5d.dev/golang@sha256:176928588492da996e1d64b66397ca6c3e7b629d2ba3e333f2b0ae10bffe637b AS build

ARG TAG="v1.0.1"

RUN set -eux; \
    apk add --no-cache git; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/greboid/greboid.com /go/src/github.com/greboid/greboid.com; \
    cd /go/src/github.com/greboid/greboid.com; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/greboid/greboid.com/main;

FROM reg.g5d.dev/base@sha256:9eb1693b711c4282b96c772aecfb3fe764a384ed907fb24f80831830e90678c8

COPY --from=build /go/src/github.com/greboid/greboid.com/main /greboid.com
EXPOSE 8080
ENTRYPOINT ["/greboid.com"]
