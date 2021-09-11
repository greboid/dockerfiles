# Generated from https://github.com/csmith/dockerfiles/blob/master/webhooked/Dockerfile.gotpl
# BOM: {"github:csmith/webhooked":"v1.0.0","image:base":"sha256:9eb1693b711c4282b96c772aecfb3fe764a384ed907fb24f80831830e90678c8","image:golang":"sha256:176928588492da996e1d64b66397ca6c3e7b629d2ba3e333f2b0ae10bffe637b"}

FROM reg.g5d.dev/golang@sha256:176928588492da996e1d64b66397ca6c3e7b629d2ba3e333f2b0ae10bffe637b AS build

ARG TAG="v1.0.0"

RUN set -eux; \
    apk add --no-cache git; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/csmith/webhooked /go/src/github.com/csmith/webhooked; \
    cd /go/src/github.com/csmith/webhooked; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    go run github.com/google/go-licenses@latest save ./... --save_path=/notices; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/csmith/webhooked/main;

FROM reg.g5d.dev/base@sha256:9eb1693b711c4282b96c772aecfb3fe764a384ed907fb24f80831830e90678c8

COPY --from=build /go/src/github.com/csmith/webhooked/main /webhooked
COPY --from=build /notices /notices

ENTRYPOINT ["/webhooked"]
