# Generated from https://github.com/greboid/dockerfiles/blob/master/soju/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r6","apk:busybox":"1.35.0-r17","apk:ca-certificates":"20220614-r0","apk:expat":"2.5.0-r0","apk:git":"2.36.3-r0","apk:libcrypto1.1":"1.1.1s-r0","apk:libcurl":"7.83.1-r4","apk:libssl1.1":"1.1.1s-r0","apk:musl":"1.2.3-r2","apk:nghttp2-libs":"1.47.0-r0","apk:patch":"2.7.6-r7","apk:pcre2":"10.40-r0","apk:zlib":"1.2.12-r3","github:csmith/certwrapper":"v4.1.0","github:emersion/soju":"v0.5.2","image:base":"94a9a1a6a1fa72a3d617c43e69f1a125e82e49d5381b3ea37dfaafc2a21ef3ee","image:golang":"32d7e7bbe25f01395b05f333e431a894887e1601281e923cc712d85b7e178cd7"}

FROM reg.g5d.dev/golang@sha256:32d7e7bbe25f01395b05f333e431a894887e1601281e923cc712d85b7e178cd7 AS build

ARG SOJUTAG="v0.5.2"
ARG CERTTAG="v4.1.0"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r6 \
        busybox=1.35.0-r17 \
        ca-certificates=20220614-r0 \
        expat=2.5.0-r0 \
        git=2.36.3-r0 \
        libcrypto1.1=1.1.1s-r0 \
        libcurl=7.83.1-r4 \
        libssl1.1=1.1.1s-r0 \
        musl=1.2.3-r2 \
        nghttp2-libs=1.47.0-r0 \
        patch=2.7.6-r7 \
        pcre2=10.40-r0 \
        zlib=1.2.12-r3 \
        ; \
    git -c advice.detachedHead=false clone --depth=1 -b $SOJUTAG --single-branch https://github.com/emersion/soju /go/src/github.com/emersion/soju; \
    cd /go/src/github.com/emersion/soju; \
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o soju ./cmd/soju; \
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o sojuctl ./cmd/sojuctl; \
    git -c advice.detachedHead=false clone --depth=1 -b $CERTTAG --single-branch https://github.com/csmith/certwrapper /go/src/github.com/csmith/certwrapper; \
    cd /go/src/github.com/csmith/certwrapper;\
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -tags lego_httpreq -o certwrapper .; \
    go-licenses save ./... --save_path=/notices; \
    mkdir /data

FROM reg.g5d.dev/base@sha256:94a9a1a6a1fa72a3d617c43e69f1a125e82e49d5381b3ea37dfaafc2a21ef3ee

COPY --from=build /go/src/github.com/emersion/soju/soju /soju
COPY --from=build /go/src/github.com/emersion/soju/sojuctl /sojuctl
COPY --from=build /go/src/github.com/csmith/certwrapper/certwrapper /certwrapper
COPY --from=build /notices /notices
COPY --from=build --chown=65535:65535 /data /data
ENTRYPOINT ["/certwrapper", "/soju"]
