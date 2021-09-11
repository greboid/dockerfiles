# Generated from https://github.com/csmith/dockerfiles/blob/master/watchtower/Dockerfile.gotpl
# BOM: {"github:containrrr/watchtower":"v1.3.0","image:baseroot":"sha256:73cd5d7b8df5d4846f076ca7c8181d4e5d7a60bb4ba4e9a66fa55e75a57d2e23","image:golang":"sha256:176928588492da996e1d64b66397ca6c3e7b629d2ba3e333f2b0ae10bffe637b"}

FROM reg.g5d.dev/golang@sha256:176928588492da996e1d64b66397ca6c3e7b629d2ba3e333f2b0ae10bffe637b AS build

ARG TAG="v1.3.0"

RUN set -eux; \
    apk add --no-cache git; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/containrrr/watchtower /go/src/github.com/containrrr/watchtower; \
    cd /go/src/github.com/containrrr/watchtower; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    #go run github.com/google/go-licenses@latest save ./... --save_path=/notices; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/containrrr/watchtower/main;

FROM reg.g5d.dev/baseroot@sha256:73cd5d7b8df5d4846f076ca7c8181d4e5d7a60bb4ba4e9a66fa55e75a57d2e23

COPY --from=build /go/src/github.com/containrrr/watchtower/main /watchtower
#COPY --from=build /notices /notices
ENTRYPOINT ["/watchtower"]
