# Generated from https://github.com/greboid/dockerfiles/blob/master/soju/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r6","apk:busybox":"1.35.0-r16","apk:ca-certificates":"20220614-r0","apk:expat":"2.4.8-r0","apk:git":"2.36.2-r0","apk:libcrypto1.1":"1.1.1q-r0","apk:libcurl":"7.83.1-r2","apk:libssl1.1":"1.1.1q-r0","apk:musl":"1.2.3-r0","apk:nghttp2-libs":"1.47.0-r0","apk:patch":"2.7.6-r7","apk:pcre2":"10.40-r0","apk:zlib":"1.2.12-r1","github:csmith/certwrapper":"v4.1.0","github:emersion/soju":"v0.5.1","image:base":"2bff9c8a562d79bf78edabbc323b647a7cccb691b2dd71f2c9c3a83df6c99f97","image:golang":"b3d0fc5d2f42fe670184b994966b93886611564c07f4dd3cd38765aca577466a"}

FROM reg.g5d.dev/golang@sha256:b3d0fc5d2f42fe670184b994966b93886611564c07f4dd3cd38765aca577466a AS build

ARG SOJUTAG="v0.5.1"
ARG CERTTAG="v4.1.0"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r6 \
        busybox=1.35.0-r16 \
        ca-certificates=20220614-r0 \
        expat=2.4.8-r0 \
        git=2.36.2-r0 \
        libcrypto1.1=1.1.1q-r0 \
        libcurl=7.83.1-r2 \
        libssl1.1=1.1.1q-r0 \
        musl=1.2.3-r0 \
        nghttp2-libs=1.47.0-r0 \
        patch=2.7.6-r7 \
        pcre2=10.40-r0 \
        zlib=1.2.12-r1 \
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

FROM reg.g5d.dev/base@sha256:2bff9c8a562d79bf78edabbc323b647a7cccb691b2dd71f2c9c3a83df6c99f97

COPY --from=build /go/src/github.com/emersion/soju/soju /soju
COPY --from=build /go/src/github.com/emersion/soju/sojuctl /sojuctl
COPY --from=build /go/src/github.com/csmith/certwrapper/certwrapper /certwrapper
COPY --from=build /notices /notices
COPY --from=build --chown=65535:65535 /data /data
ENTRYPOINT ["/certwrapper", "/soju"]
