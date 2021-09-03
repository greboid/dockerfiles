# Generated from https://github.com/csmith/dockerfiles/blob/master/distribution/Dockerfile.gotpl
# BOM: {"image:base":"sha256:1cc5b31fea33445d7b10537f4e5c2682fdc72e806b7c92dbdd3930a88fed86f0","image:golang":"sha256:b49ff0d1a458cc16f0a773ffffa321568e689f2abd2250a181b95c0cfb0c3567"}

FROM reg.g5d.dev/golang@sha256:b49ff0d1a458cc16f0a773ffffa321568e689f2abd2250a181b95c0cfb0c3567 AS build

ARG TAG="v2.7.1"

RUN set -eux; \
    apk add --no-cache make git file; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/distribution/distribution /go/src/github.com/docker/distribution; \
    cd /go/src/github.com/docker/distribution; \
    BUILDTAGS="include_oss include_gcs" GO111MODULE=off CGO_ENABLED=0 make bin/registry; \
    # Clobber all the timestamps to make the build more reproducible
    touch --date=@0 /go/src/github.com/docker/distribution/cmd/registry/config-dev.yml /go/src/github.com/docker/distribution/bin/registry;

FROM reg.g5d.dev/base@sha256:1cc5b31fea33445d7b10537f4e5c2682fdc72e806b7c92dbdd3930a88fed86f0

COPY --from=build /go/src/github.com/docker/distribution/cmd/registry/config-dev.yml /etc/docker/registry/config.yml
COPY --from=build /go/src/github.com/docker/distribution/bin/registry /bin/registry
EXPOSE 5000
VOLUME ["/var/lib/registry"]
ENTRYPOINT ["/bin/registry"]
CMD ["serve", "/etc/docker/registry/config.yml"]
