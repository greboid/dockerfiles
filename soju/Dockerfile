# Generated from https://github.com/csmith/dockerfiles/blob/master/soju/Dockerfile.gotpl
# BOM: {"github:csmith/certwrapper":"v2.0.0","github:emersion/soju":"v0.2.0","image:base":"sha256:9eb1693b711c4282b96c772aecfb3fe764a384ed907fb24f80831830e90678c8","image:golang":"sha256:176928588492da996e1d64b66397ca6c3e7b629d2ba3e333f2b0ae10bffe637b"}

FROM reg.g5d.dev/golang@sha256:176928588492da996e1d64b66397ca6c3e7b629d2ba3e333f2b0ae10bffe637b AS build

ARG SOJUTAG="v0.2.0"
ARG CERTTAG="v2.0.0"

RUN set -eux; \
    apk add --no-cache git; \
    git clone --depth=1 -b $SOJUTAG --single-branch https://github.com/emersion/soju /go/src/github.com/emersion/soju; \
    cd /go/src/github.com/emersion/soju; \
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o soju ./cmd/soju; \
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o sojuctl ./cmd/sojuctl; \
    git clone --depth=1 -b $CERTTAG --single-branch https://github.com/csmith/certwrapper /go/src/github.com/csmith/certwrapper; \
    cd /go/src/github.com/csmith/certwrapper;\
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o certwrapper .; \
    go run github.com/google/go-licenses@latest save ./... --save_path=/notices; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/emersion/soju/soju; \
    touch --date=@0 /go/src/github.com/emersion/soju/sojuctl; \
    touch --date=@0 /go/src/github.com/csmith/certwrapper/certwrapper; \
    mkdir /data

FROM reg.g5d.dev/base@sha256:9eb1693b711c4282b96c772aecfb3fe764a384ed907fb24f80831830e90678c8

COPY --from=build /go/src/github.com/emersion/soju/soju /soju
COPY --from=build /go/src/github.com/emersion/soju/sojuctl /sojuctl
COPY --from=build /go/src/github.com/csmith/certwrapper/certwrapper /certwrapper
COPY --from=build /notices /notices
COPY --from=build --chown=65535:65535 /data /data
ENTRYPOINT ["/certwrapper", "/soju"]
