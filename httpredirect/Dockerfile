# Generated from https://github.com/csmith/dockerfiles/blob/master/httpredirect/Dockerfile.gotpl
# BOM: {"github:greboid/httpredirect":"v1.0.0","image:base":"sha256:4dc61e45d55285af7ae06b1f56943e825d55793a628d085ff56be2a00a3d5039","image:golang":"sha256:092432223820de2928ae7bd05e069c52b69349e29250695f56b793e17c32d0a2"}

FROM reg.g5d.dev/golang@sha256:092432223820de2928ae7bd05e069c52b69349e29250695f56b793e17c32d0a2 AS build

ARG TAG="v1.0.0"

RUN set -eux; \
    apk add --no-cache git; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/greboid/httpredirect /go/src/github.com/greboid/httpredirect; \
    cd /go/src/github.com/greboid/httpredirect; \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main .; \
    go run github.com/google/go-licenses@latest save ./... --save_path=/notices; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/greboid/httpredirect/main;

FROM reg.g5d.dev/base@sha256:4dc61e45d55285af7ae06b1f56943e825d55793a628d085ff56be2a00a3d5039

COPY --from=build /go/src/github.com/greboid/httpredirect/main /httpredirect
COPY --from=build /notices /notices
EXPOSE 8080
ENTRYPOINT ["/httpredirect"]
