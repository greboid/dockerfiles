# Generated from https://github.com/greboid/dockerfiles/blob/master/registryauth/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r4","apk:git":"2.38.4-r1","apk:libcrypto3":"3.0.8-r0","apk:libcurl":"7.88.1-r1","apk:libexpat":"2.5.0-r0","apk:libssl3":"3.0.8-r0","apk:musl":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:pcre2":"10.42-r0","apk:zlib":"1.2.13-r0","github:greboid/registryauth":"v1.0.20","image:base":"ffd04efedafc54f5392c35a1045604c12a4329e5780407d5fe08dfe495d411b1","image:golang":"6bd1454d3b9011ba0e4f9e1ddeb4a0f2000ae065f69d59f708eada260127a8a2"}

FROM reg.g5d.dev/golang@sha256:6bd1454d3b9011ba0e4f9e1ddeb4a0f2000ae065f69d59f708eada260127a8a2 AS build

ARG TAG="v1.0.20"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r9\
        busybox=1.35.0-r29\
        busybox-binsh=1.35.0-r29\
        ca-certificates=20220614-r4\
        git=2.38.4-r1\
        libcrypto3=3.0.8-r0\
        libcurl=7.88.1-r1\
        libexpat=2.5.0-r0\
        libssl3=3.0.8-r0\
        musl=1.2.3-r4\
        nghttp2-libs=1.51.0-r0\
        pcre2=10.42-r0\
        zlib=1.2.13-r0\
        ; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/greboid/registryauth /go/src/github.com/greboid/registryauth; \
    cd /go/src/github.com/greboid/registryauth; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main ./cmd/authedreg; \
    go-licenses save ./... --save_path=/notices; \
    mkdir /data

FROM reg.g5d.dev/base@sha256:ffd04efedafc54f5392c35a1045604c12a4329e5780407d5fe08dfe495d411b1

COPY --from=build /go/src/github.com/greboid/registryauth/main /authedreg
COPY --from=build --chown=65532:65532 /data /data
COPY --from=build /notices /notices
EXPOSE 8080
ENTRYPOINT ["/authedreg"]
