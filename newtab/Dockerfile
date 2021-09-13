# Generated from https://github.com/csmith/dockerfiles/blob/master/newtab/Dockerfile.gotpl
# BOM: {"github:greboid/newtab":"v1.0.0","image:base":"sha256:9eb1693b711c4282b96c772aecfb3fe764a384ed907fb24f80831830e90678c8","image:golang":"sha256:176928588492da996e1d64b66397ca6c3e7b629d2ba3e333f2b0ae10bffe637b"}

FROM reg.g5d.dev/golang@sha256:176928588492da996e1d64b66397ca6c3e7b629d2ba3e333f2b0ae10bffe637b AS build

ARG TAG="v1.0.0"

RUN set -eux; \
    apk add --no-cache git; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/greboid/newtab /go/src/github.com/greboid/newtab; \
    cd /go/src/github.com/greboid/newtab; \
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    go run github.com/google/go-licenses@latest save ./... --save_path=/notices; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/greboid/newtab/main;

FROM reg.g5d.dev/base@sha256:9eb1693b711c4282b96c772aecfb3fe764a384ed907fb24f80831830e90678c8

COPY --from=build /go/src/github.com/greboid/newtab/main /newtab
COPY --from=build /notices /notices
EXPOSE 8080
ENTRYPOINT ["/newtab"]
