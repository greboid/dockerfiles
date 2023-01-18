# Generated from https://github.com/greboid/dockerfiles/blob/master/githubmirror/Dockerfile.gotpl
# BOM: {"apk:brotli-libs":"1.0.9-r9","apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r4","apk:git":"2.38.3-r1","apk:libcrypto3":"3.0.7-r2","apk:libcurl":"7.87.0-r1","apk:libexpat":"2.5.0-r0","apk:libssl3":"3.0.7-r2","apk:musl":"1.2.3-r4","apk:nghttp2-libs":"1.51.0-r0","apk:pcre2":"10.42-r0","apk:zlib":"1.2.13-r0","github:greboid/githubmirror":"v1.0.1","image:base":"ffd04efedafc54f5392c35a1045604c12a4329e5780407d5fe08dfe495d411b1","image:golang":"62336f81c6c2c3ecdf35de9cf62c7ad3e84cebad54841ad004c734da9ca56bbe"}

FROM reg.g5d.dev/golang@sha256:62336f81c6c2c3ecdf35de9cf62c7ad3e84cebad54841ad004c734da9ca56bbe AS build

ARG TAG="v1.0.1"

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
    git clone --depth=1 -b $TAG --single-branch https://github.com/greboid/githubmirror /go/src/github.com/greboid/githubmirror; \
    cd /go/src/github.com/greboid/githubmirror; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    go-licenses save ./... --save_path=/notices;

FROM reg.g5d.dev/base@sha256:ffd04efedafc54f5392c35a1045604c12a4329e5780407d5fe08dfe495d411b1

COPY --from=build /go/src/github.com/greboid/githubmirror/main /githubmirror
COPY --from=build /notices /notices
EXPOSE 8080
ENTRYPOINT ["/githubmirror"]
