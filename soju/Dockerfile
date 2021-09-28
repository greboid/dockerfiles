# Generated from https://github.com/csmith/dockerfiles/blob/master/soju/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r5","apk:busybox":"1.33.1-r3","apk:ca-certificates":"20191127-r5","apk:expat":"2.4.1-r0","apk:git":"2.32.0-r0","apk:libcrypto1.1":"1.1.1l-r0","apk:libcurl":"7.79.1-r0","apk:libssl1.1":"1.1.1l-r0","apk:musl":"1.2.2-r3","apk:nghttp2-libs":"1.43.0-r0","apk:patch":"2.7.6-r7","apk:pcre2":"10.36-r0","apk:zlib":"1.2.11-r3","github:csmith/certwrapper":"v3.1.2","github:emersion/soju":"v0.2.1","image:base":"sha256:712fc3471f4bc3f997c6fde284cde3b351acdc88a5904fc8488ee557d3e67904","image:golang":"sha256:5ee724772fe7a6042d23044146b8db2b01943f2a24a3d9c49a792d7cb9051a79"}

FROM reg.g5d.dev/golang@sha256:5ee724772fe7a6042d23044146b8db2b01943f2a24a3d9c49a792d7cb9051a79 AS build

ARG SOJUTAG="v0.2.1"
ARG CERTTAG="v3.1.2"

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
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -tags httpreq -o certwrapper .; \
    go run github.com/google/go-licenses@latest save ./... --save_path=/notices; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/emersion/soju/soju; \
    touch --date=@0 /go/src/github.com/emersion/soju/sojuctl; \
    touch --date=@0 /go/src/github.com/csmith/certwrapper/certwrapper; \
    mkdir /data

FROM reg.g5d.dev/base@sha256:712fc3471f4bc3f997c6fde284cde3b351acdc88a5904fc8488ee557d3e67904

COPY --from=build /go/src/github.com/emersion/soju/soju /soju
COPY --from=build /go/src/github.com/emersion/soju/sojuctl /sojuctl
COPY --from=build /go/src/github.com/csmith/certwrapper/certwrapper /certwrapper
COPY --from=build /notices /notices
COPY --from=build --chown=65535:65535 /data /data
ENTRYPOINT ["/certwrapper", "/soju"]
