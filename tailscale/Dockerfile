# Generated from https://github.com/greboid/dockerfiles/blob/master/tailscale/Dockerfile.gotpl
# BOM: {"github:tailscale/tailscale":"v1.40.1","image:base":"4ddff4468810fe9d9dc3285917d2923c41ae9fb6d1fa2b91c3f1d460cb94427a","image:golang":"866a26e8691570c5e39e43030b288dde9b22387f36f3b58d659af14fcbbf9477"}

FROM reg.g5d.dev/golang@sha256:866a26e8691570c5e39e43030b288dde9b22387f36f3b58d659af14fcbbf9477 AS build

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

FROM reg.g5d.dev/base@sha256:4ddff4468810fe9d9dc3285917d2923c41ae9fb6d1fa2b91c3f1d460cb94427a

COPY --from=build --chown=65532:65532 /rootfs/ /

ENTRYPOINT ["/tailscaled", "--tun", "userspace-networking", "--statedir", "/state/"]
