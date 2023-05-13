# Generated from https://github.com/greboid/dockerfiles/blob/master/tailscale/Dockerfile.gotpl
# BOM: {"github:tailscale/tailscale":"v1.40.1","image:base":"a98010a83248020eabd68f86ab1c1711d43693c5bb2e0db86e227b37af2e7ec1","image:golang":"89d31e144ad23468190029aa512882c2b3709a73bc9360a55a6612f1fa0a8629"}

FROM reg.g5d.dev/golang@sha256:89d31e144ad23468190029aa512882c2b3709a73bc9360a55a6612f1fa0a8629 AS build

ARG TAG="v1.40.1"

RUN set -eux; \
    git clone --depth=1 -b $TAG --single-branch https://github.com/tailscale/tailscale /src; \
    cd /src; \
    VERS=$(echo $TAG | tr -d 'v'); \
    VERL=$VERS-t$(git rev-parse HEAD | cut -c1-9); \
    VERC=$(git rev-parse HEAD); \
    CGO_ENABLED=0 go build -trimpath -ldflags "-extldflags \"-static\" -buildid= -s -w -X tailscale.com/version.Long=$VERL -X tailscale.com/version.Short=$VERS -X tailscale.com/version.GitCommit=$VERC" ./cmd/tailscale; \
    CGO_ENABLED=0 go build -trimpath -ldflags "-extldflags \"-static\" -buildid= -s -w -X tailscale.com/version.Long=$VERL -X tailscale.com/version.Short=$VERS -X tailscale.com/version.GitCommit=$VERC" ./cmd/tailscaled; \
    mkdir -p /rootfs/state /rootfs/var/run/tailscale; \
    cp /src/tailscaled /src/tailscale /rootfs/;

FROM reg.g5d.dev/base@sha256:a98010a83248020eabd68f86ab1c1711d43693c5bb2e0db86e227b37af2e7ec1

COPY --from=build --chown=65532:65532 /rootfs/ /

ENTRYPOINT ["/tailscaled", "--tun", "userspace-networking", "--statedir", "/state/"]
