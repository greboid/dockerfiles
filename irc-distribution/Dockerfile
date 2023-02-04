# Generated from https://github.com/greboid/dockerfiles/blob/master/irc-distribution/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r4","apk:git":"2.38.3-r1","apk:libcrypto3":"3.0.7-r2","apk:libcurl":"7.87.0-r1","apk:libexpat":"2.5.0-r0","apk:libssl3":"3.0.7-r2","apk:musl":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:pcre2":"10.42-r0","apk:zlib":"1.2.13-r0","github:csmith/irc-distribution":"v1.0.0","image:base":"ffd04efedafc54f5392c35a1045604c12a4329e5780407d5fe08dfe495d411b1","image:golang":"ada409e9ecb72011eb4bcddfa2ccbb9cb08804c14a8dd1e440f398e93b9285b9"}

FROM reg.g5d.dev/golang@sha256:ada409e9ecb72011eb4bcddfa2ccbb9cb08804c14a8dd1e440f398e93b9285b9 AS build

ARG TAG="v1.0.0"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r9\
        busybox=1.35.0-r29\
        busybox-binsh=1.35.0-r29\
        ca-certificates=20220614-r4\
        git=2.38.3-r1\
        libcrypto3=3.0.7-r2\
        libcurl=7.87.0-r1\
        libexpat=2.5.0-r0\
        libssl3=3.0.7-r2\
        musl=1.2.3-r4\
        nghttp2-libs=1.51.0-r0\
        pcre2=10.42-r0\
        zlib=1.2.13-r0\
        ; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/csmith/irc-distribution /go/src/github.com/csmith/irc-distribution; \
    cd /go/src/github.com/csmith/irc-distribution; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main ./cmd/distribution; \
    go-licenses save ./... --save_path=/notices;

FROM reg.g5d.dev/base@sha256:ffd04efedafc54f5392c35a1045604c12a4329e5780407d5fe08dfe495d411b1

COPY --from=build /go/src/github.com/csmith/irc-distribution/main /irc-distribution
COPY --from=build /notices /notices
ENTRYPOINT ["/irc-distribution"]
