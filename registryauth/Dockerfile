# Generated from https://github.com/greboid/dockerfiles/blob/master/registryauth/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20230506-r0","apk:git":"2.38.5-r0","apk:libcrypto3":"3.0.8-r4","apk:libcurl":"8.0.1-r0","apk:libexpat":"2.5.0-r0","apk:libssl3":"3.0.8-r4","apk:musl":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:pcre2":"10.42-r0","apk:zlib":"1.2.13-r0","github:greboid/registryauth":"v1.0.20","image:base":"2cf8662a916097fbef5fc9aa384be213a8bf229277931807b092431c9c77e90f","image:golang":"5bb7e1c27710c2d0390e6b1456c52ee9d1702cb5e6c4b54cad653b6d1cf7deb4"}

FROM reg.g5d.dev/golang@sha256:5bb7e1c27710c2d0390e6b1456c52ee9d1702cb5e6c4b54cad653b6d1cf7deb4 AS build

ARG TAG="v1.0.20"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r9\
        busybox=1.35.0-r29\
        busybox-binsh=1.35.0-r29\
        ca-certificates=20230506-r0\
        git=2.38.5-r0\
        libcrypto3=3.0.8-r4\
        libcurl=8.0.1-r0\
        libexpat=2.5.0-r0\
        libssl3=3.0.8-r4\
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

FROM reg.g5d.dev/base@sha256:2cf8662a916097fbef5fc9aa384be213a8bf229277931807b092431c9c77e90f

COPY --from=build /go/src/github.com/greboid/registryauth/main /authedreg
COPY --from=build --chown=65532:65532 /data /data
COPY --from=build /notices /notices
EXPOSE 8080
ENTRYPOINT ["/authedreg"]
