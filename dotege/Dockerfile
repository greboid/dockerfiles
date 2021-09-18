# Generated from https://github.com/csmith/dockerfiles/blob/master/dotege/Dockerfile.gotpl
# BOM: {"github:csmith/dotege":"v1.0.0","image:baseroot":"sha256:da3ab01f29ec0054e06e485d216c219a62efa78691f20e2cd67d4ead3688ab05","image:golang":"sha256:092432223820de2928ae7bd05e069c52b69349e29250695f56b793e17c32d0a2"}

FROM reg.g5d.dev/golang@sha256:092432223820de2928ae7bd05e069c52b69349e29250695f56b793e17c32d0a2 AS build

ARG TAG="v1.0.0"

RUN set -eux; \
    apk add --no-cache git; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/csmith/dotege /go/src/github.com/csmith/dotege; \
    cd /go/src/github.com/csmith/dotege; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    go run github.com/google/go-licenses@latest save ./... --save_path=/notices; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/csmith/dotege/main;

FROM reg.g5d.dev/baseroot@sha256:da3ab01f29ec0054e06e485d216c219a62efa78691f20e2cd67d4ead3688ab05

COPY --from=build /go/src/github.com/csmith/dotege/main /dotege
COPY --from=build /notices /notices
COPY --from=build /go/src/github.com/csmith/dotege/templates /templates

VOLUME /data/config
VOLUME /data/certs
VOLUME /data/output

ENTRYPOINT ["/dotege"]
