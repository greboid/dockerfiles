# Generated from https://github.com/csmith/dockerfiles/blob/master/miniflux/Dockerfile.gotpl
# BOM: {"github:miniflux/v2":"2.0.32","image:base":"sha256:f7f27db7afb58bae23ad902072228f1b090b878af303d9f63bd2c1526b9b4f53","image:golang":"sha256:beae544bfe746af2dec32e19106ea3ce7e346dde3c9692bb1a82c70be9b636e4"}

FROM reg.g5d.dev/golang@sha256:beae544bfe746af2dec32e19106ea3ce7e346dde3c9692bb1a82c70be9b636e4 AS build

ARG TAG="2.0.32"

RUN set -eux; \
    apk add --no-cache git; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/miniflux/v2 /go/src/github.com/miniflux/v2; \
    cd /go/src/github.com/miniflux/v2; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    go run github.com/google/go-licenses@latest save ./... --save_path=/notices; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/miniflux/v2/main;

FROM reg.g5d.dev/base@sha256:f7f27db7afb58bae23ad902072228f1b090b878af303d9f63bd2c1526b9b4f53

COPY --from=build /go/src/github.com/miniflux/v2/main /miniflux
COPY --from=build /notices /notices
ENTRYPOINT ["/miniflux"]
