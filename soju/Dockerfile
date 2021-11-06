# Generated from https://github.com/greboid/dockerfiles/blob/master/soju/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.33.1-r3","apk:ca-certificates":"20191127-r5","apk:expat":"2.4.1-r0","apk:git":"2.32.0-r0","apk:libcrypto1.1":"1.1.1l-r0","apk:libcurl":"7.79.1-r0","apk:libssl1.1":"1.1.1l-r0","apk:musl":"1.2.2-r3","apk:nghttp2-libs":"1.43.0-r0","apk:patch":"2.7.6-r7","apk:pcre2":"10.36-r0","apk:zlib":"1.2.11-r3","github:csmith/certwrapper":"v4.0.0","github:emersion/soju":"v0.2.2","image:base":"f25b9dc29d96dee9ff3919d93729ba9105802a4ee8a15b69e654877fa74e0af2","image:golang":"fbd91b4cd6bfd5cfbe7b53d323e456ae0b4ae19d4b1638f8bd731355ad551b74"}

FROM reg.g5d.dev/golang@sha256:fbd91b4cd6bfd5cfbe7b53d323e456ae0b4ae19d4b1638f8bd731355ad551b74 AS build

ARG SOJUTAG="v0.2.2"
ARG CERTTAG="v4.0.0"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r5 \
        busybox=1.33.1-r3 \
        ca-certificates=20191127-r5 \
        expat=2.4.1-r0 \
        git=2.32.0-r0 \
        libcrypto1.1=1.1.1l-r0 \
        libcurl=7.79.1-r0 \
        libssl1.1=1.1.1l-r0 \
        musl=1.2.2-r3 \
        nghttp2-libs=1.43.0-r0 \
        patch=2.7.6-r7 \
        pcre2=10.36-r0 \
        zlib=1.2.11-r3 \
        ; \
    git clone --depth=1 -b $SOJUTAG --single-branch https://github.com/emersion/soju /go/src/github.com/emersion/soju; \
    cd /go/src/github.com/emersion/soju; \
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o soju ./cmd/soju; \
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o sojuctl ./cmd/sojuctl; \
    git clone --depth=1 -b $CERTTAG --single-branch https://github.com/csmith/certwrapper /go/src/github.com/csmith/certwrapper; \
    cd /go/src/github.com/csmith/certwrapper;\
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -tags lego_httpreq -o certwrapper .; \
    go run github.com/google/go-licenses@latest save ./... --save_path=/notices; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/emersion/soju/soju; \
    touch --date=@0 /go/src/github.com/emersion/soju/sojuctl; \
    touch --date=@0 /go/src/github.com/csmith/certwrapper/certwrapper; \
    mkdir /data

FROM reg.g5d.dev/base@sha256:f25b9dc29d96dee9ff3919d93729ba9105802a4ee8a15b69e654877fa74e0af2

COPY --from=build /go/src/github.com/emersion/soju/soju /soju
COPY --from=build /go/src/github.com/emersion/soju/sojuctl /sojuctl
COPY --from=build /go/src/github.com/csmith/certwrapper/certwrapper /certwrapper
COPY --from=build /notices /notices
COPY --from=build --chown=65535:65535 /data /data
ENTRYPOINT ["/certwrapper", "/soju"]
