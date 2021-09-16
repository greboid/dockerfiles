# Generated from https://github.com/csmith/dockerfiles/blob/master/dotege/Dockerfile.gotpl
# BOM: {"github:csmith/dotege":"v1.0.0","image:baseroot":"sha256:73cd5d7b8df5d4846f076ca7c8181d4e5d7a60bb4ba4e9a66fa55e75a57d2e23","image:golang":"sha256:c1d9d794a310682b45fbdaad1dbe94ee3773ce218c3c0fcdff6b499cc87c6f4c"}

FROM reg.g5d.dev/golang@sha256:c1d9d794a310682b45fbdaad1dbe94ee3773ce218c3c0fcdff6b499cc87c6f4c AS build

ARG TAG="v1.0.0"

RUN set -eux; \
    apk add --no-cache git; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/csmith/dotege /go/src/github.com/csmith/dotege; \
    cd /go/src/github.com/csmith/dotege; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    go run github.com/google/go-licenses@latest save ./... --save_path=/notices; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/csmith/dotege/main;

FROM reg.g5d.dev/baseroot@sha256:73cd5d7b8df5d4846f076ca7c8181d4e5d7a60bb4ba4e9a66fa55e75a57d2e23

COPY --from=build /go/src/github.com/csmith/dotege/main /dotege
COPY --from=build /notices /notices
COPY --from=build /go/src/github.com/csmith/dotege/templates /templates

VOLUME /data/config
VOLUME /data/certs
VOLUME /data/output

ENTRYPOINT ["/dotege"]
