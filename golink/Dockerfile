# Generated from https://github.com/greboid/dockerfiles/blob/master/golink/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r2","apk:git":"2.38.2-r0","apk:libcrypto3":"3.0.7-r0","apk:libcurl":"7.86.0-r1","apk:libexpat":"2.5.0-r0","apk:libssl3":"3.0.7-r0","apk:musl":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:pcre2":"10.42-r0","apk:zlib":"1.2.13-r0","github:tailscale/tailscale":"v1.34.1","image:base":"ffd04efedafc54f5392c35a1045604c12a4329e5780407d5fe08dfe495d411b1","image:golang":"a5b8eabae8594fe0273fb26b22fe43cd76fc8b0556133fc79554464c2a5a2797"}

FROM reg.g5d.dev/golang@sha256:a5b8eabae8594fe0273fb26b22fe43cd76fc8b0556133fc79554464c2a5a2797 AS build

ARG TAG=5eb91a2ccfa4d2a3ebbcd8b3048ac54f10fc0731
ARG TS=v1.34.1
RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r9\
        busybox=1.35.0-r29\
        busybox-binsh=1.35.0-r29\
        ca-certificates=20220614-r2\
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
    git clone -b main --single-branch https://github.com/tailscale/golink /src; \
    cd /src; \
    git checkout $TAG; \
#   git clone --depth=1 -b $TAG --single-branch https://github.com/tailscale/golink /src; \
    cd /src; \
    #This is a bit horrible, but I cant think of a nicer way to update the dependency and its tree
    go get -u tailscale.com@$TS; \
    go mod tidy; \
    CGO_ENABLED=0 GOOS=linux go build -a -trimpath -ldflags="-extldflags \"-static\" -buildid= -s -w" -o main ./cmd/golink; \
    go-licenses save ./... --save_path=/notices --ignore modernc.org/mathutil;

FROM reg.g5d.dev/base@sha256:ffd04efedafc54f5392c35a1045604c12a4329e5780407d5fe08dfe495d411b1

COPY --from=build /src/main /golink
COPY --from=build /notices /notices
EXPOSE 8080
ENTRYPOINT ["/golink"]
CMD ["--sqlitedb", "/home/nonroot/golink.db"]
