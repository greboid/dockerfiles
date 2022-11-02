# Generated from https://github.com/greboid/dockerfiles/blob/master/githubmirror/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r6","apk:busybox":"1.35.0-r17","apk:ca-certificates":"20220614-r0","apk:expat":"2.5.0-r0","apk:git":"2.36.3-r0","apk:libcrypto1.1":"1.1.1s-r0","apk:libcurl":"7.83.1-r4","apk:libssl1.1":"1.1.1s-r0","apk:musl":"1.2.3-r0","apk:nghttp2-libs":"1.47.0-r0","apk:pcre2":"10.40-r0","apk:zlib":"1.2.12-r3","github:greboid/githubmirror":"v1.0.1","image:base":"8b908b9450bc130ede6eb6f40f038563b8e35d64c8ac566c9ca647efa349bf2f","image:golang":"cdb40b219137c0fc711109b143ebbe41ee41d8b3f2e8aeac678412174c237941"}

FROM reg.g5d.dev/golang@sha256:cdb40b219137c0fc711109b143ebbe41ee41d8b3f2e8aeac678412174c237941 AS build

ARG TAG="v1.0.1"

RUN set -eux; \
    apk add --no-cache \
        brotli-libs=1.0.9-r6\
        busybox=1.35.0-r17\
        ca-certificates=20220614-r0\
        expat=2.5.0-r0\
        git=2.36.3-r0\
        libcrypto1.1=1.1.1s-r0\
        libcurl=7.83.1-r4\
        libssl1.1=1.1.1s-r0\
        musl=1.2.3-r0\
        nghttp2-libs=1.47.0-r0\
        pcre2=10.40-r0\
        zlib=1.2.12-r3\
        ; \
    git -c advice.detachedHead=false clone --depth=1 -b $TAG --single-branch https://github.com/greboid/githubmirror /go/src/github.com/greboid/githubmirror; \
    cd /go/src/github.com/greboid/githubmirror; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    go-licenses save ./... --save_path=/notices;

FROM reg.g5d.dev/base@sha256:8b908b9450bc130ede6eb6f40f038563b8e35d64c8ac566c9ca647efa349bf2f

COPY --from=build /go/src/github.com/greboid/githubmirror/main /githubmirror
COPY --from=build /notices /notices
EXPOSE 8080
ENTRYPOINT ["/githubmirror"]
