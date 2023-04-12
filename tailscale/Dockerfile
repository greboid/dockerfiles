# Generated from https://github.com/greboid/dockerfiles/blob/master/tailscale/Dockerfile.gotpl
# BOM: {"github:tailscale/tailscale":"v1.38.4","image:base":"ef8b65fda5d776551a87f970b02ebd1aa67f98c60b68637d910429a4e18be9b9","image:golang":"20b7658093fc4e7748edddd67a45e19fe467d846b9fcf3ceaa37a5f84f672f14"}

FROM reg.g5d.dev/golang@sha256:20b7658093fc4e7748edddd67a45e19fe467d846b9fcf3ceaa37a5f84f672f14 AS build

ARG TAG="v1.38.4"

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

FROM reg.g5d.dev/base@sha256:ef8b65fda5d776551a87f970b02ebd1aa67f98c60b68637d910429a4e18be9b9

COPY --from=build --chown=65532:65532 /rootfs/ /

ENTRYPOINT ["/tailscaled", "--tun", "userspace-networking", "--statedir", "/state/"]
