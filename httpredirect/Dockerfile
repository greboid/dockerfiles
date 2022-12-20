# Generated from https://github.com/greboid/dockerfiles/blob/master/httpredirect/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r3","apk:git":"2.38.2-r0","apk:libcrypto3":"3.0.7-r0","apk:libcurl":"7.86.0-r1","apk:libexpat":"2.5.0-r0","apk:libssl3":"3.0.7-r0","apk:musl":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:pcre2":"10.42-r0","apk:zlib":"1.2.13-r0","github:greboid/httpredirect":"v1.0.1","image:base":"ffd04efedafc54f5392c35a1045604c12a4329e5780407d5fe08dfe495d411b1","image:golang":"079d27d607d31c2a0f0531017af00dc1d9a24dfaa5d22b56eafe1646530eb813"}

FROM reg.g5d.dev/golang@sha256:079d27d607d31c2a0f0531017af00dc1d9a24dfaa5d22b56eafe1646530eb813 AS build

ARG TAG="v1.0.1"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r9\
        busybox=1.35.0-r29\
        busybox-binsh=1.35.0-r29\
        ca-certificates=20220614-r3\
        git=2.38.2-r0\
        libcrypto3=3.0.7-r0\
        libcurl=7.86.0-r1\
        libexpat=2.5.0-r0\
        libssl3=3.0.7-r0\
        musl=1.2.3-r4\
        nghttp2-libs=1.51.0-r0\
        pcre2=10.42-r0\
        zlib=1.2.13-r0\
        ; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/greboid/httpredirect /go/src/github.com/greboid/httpredirect; \
    cd /go/src/github.com/greboid/httpredirect; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    go-licenses save ./... --save_path=/notices;

FROM reg.g5d.dev/base@sha256:ffd04efedafc54f5392c35a1045604c12a4329e5780407d5fe08dfe495d411b1

COPY --from=build /go/src/github.com/greboid/httpredirect/main /httpredirect
COPY --from=build /notices /notices
EXPOSE 8080
ENTRYPOINT ["/httpredirect"]
