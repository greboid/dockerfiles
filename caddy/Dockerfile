# Generated from https://github.com/greboid/dockerfiles/blob/master/caddy/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r6","apk:busybox":"1.35.0-r17","apk:ca-certificates":"20220614-r0","apk:expat":"2.4.8-r0","apk:git":"2.36.2-r0","apk:libcrypto1.1":"1.1.1q-r0","apk:libcurl":"7.83.1-r3","apk:libssl1.1":"1.1.1q-r0","apk:musl":"1.2.3-r0","apk:nghttp2-libs":"1.47.0-r0","apk:pcre2":"10.40-r0","apk:zlib":"1.2.12-r3","github:caddyserver/caddy":"v2.5.2","image:base":"148dcee1d765d04d178e78cbf9e1c44b4c23af1097d0e79418893051e0e13c69","image:golang":"1d858c110c95d06e6220de01081bc03822c90354a7263adb98881f0e91675e99"}

FROM reg.g5d.dev/golang@sha256:1d858c110c95d06e6220de01081bc03822c90354a7263adb98881f0e91675e99 AS build

COPY *.patch /go

ARG REPO="caddyserver/caddy"
ARG TAG="v2.5.2"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r6\
        busybox=1.35.0-r17\
        ca-certificates=20220614-r0\
        expat=2.4.8-r0\
        git=2.36.2-r0\
        libcrypto1.1=1.1.1q-r0\
        libcurl=7.83.1-r3\
        libssl1.1=1.1.1q-r0\
        musl=1.2.3-r0\
        nghttp2-libs=1.47.0-r0\
        pcre2=10.40-r0\
        zlib=1.2.12-r3\
        ; \
    git -c advice.detachedHead=false clone --depth=1 -b $TAG --single-branch https://github.com/$REPO /go/src; \
    cd /go/src; \
    git apply --ignore-space-change --ignore-whitespace /go/removeserverheader.patch; \
    GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main ./cmd/caddy/main.go; \
    chmod +x /go/src/main;

FROM reg.g5d.dev/base@sha256:148dcee1d765d04d178e78cbf9e1c44b4c23af1097d0e79418893051e0e13c69

COPY --from=build /go/src/main /caddy
COPY CaddyLicense /notices/LICENSE
EXPOSE 8080
ENTRYPOINT ["/caddy", "run", "-config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile"]
