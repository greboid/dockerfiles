# Generated from https://github.com/greboid/dockerfiles/blob/master/dotege/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r4","apk:git":"2.38.2-r0","apk:libcrypto3":"3.0.7-r2","apk:libcurl":"7.87.0-r1","apk:libexpat":"2.5.0-r0","apk:libssl3":"3.0.7-r2","apk:musl":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:pcre2":"10.42-r0","apk:zlib":"1.2.13-r0","github:csmith/dotege":"v1.3.1","image:base":"ffd04efedafc54f5392c35a1045604c12a4329e5780407d5fe08dfe495d411b1","image:golang":"a4545095e83b65679a3e7933a43625b5fde939739fe164f929909acde3309e4f"}

FROM reg.g5d.dev/golang@sha256:a4545095e83b65679a3e7933a43625b5fde939739fe164f929909acde3309e4f AS build

ARG TAG="v1.3.1"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r9\
        busybox=1.35.0-r29\
        busybox-binsh=1.35.0-r29\
        ca-certificates=20220614-r4\
        git=2.38.2-r0\
        libcrypto3=3.0.7-r2\
        libcurl=7.87.0-r1\
        libexpat=2.5.0-r0\
        libssl3=3.0.7-r2\
        musl=1.2.3-r4\
        nghttp2-libs=1.51.0-r0\
        pcre2=10.42-r0\
        zlib=1.2.13-r0\
        ; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/csmith/dotege /go/src/github.com/csmith/dotege; \
    cd /go/src/github.com/csmith/dotege; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -tags lego_httpreq -ldflags=-buildid= -o main ./cmd/dotege; \
    go-licenses save ./... --save_path=/notices;

FROM reg.g5d.dev/base@sha256:ffd04efedafc54f5392c35a1045604c12a4329e5780407d5fe08dfe495d411b1

COPY --from=build /go/src/github.com/csmith/dotege/main /dotege
COPY --from=build /notices /notices
COPY --from=build /go/src/github.com/csmith/dotege/templates /templates

VOLUME /data/config
VOLUME /data/certs
VOLUME /data/output

ENTRYPOINT ["/dotege"]
