# Generated from https://github.com/greboid/dockerfiles/blob/master/tailscale/Containerfile.gotpl
# BOM: {"github:tailscale/tailscale":"v1.46.0","image:base":"5c298191cd12c92b8b3df64ab907f33118d5ddffa99f0c31d43c9623330c46e2","image:golang":"36190d5ec7bcb2ccd44dcffdcc6376553799520817f7654c1cf63883a285695f"}

FROM reg.g5d.dev/golang@sha256:36190d5ec7bcb2ccd44dcffdcc6376553799520817f7654c1cf63883a285695f AS build

ARG TAG="v1.46.0"

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

FROM reg.g5d.dev/base@sha256:5c298191cd12c92b8b3df64ab907f33118d5ddffa99f0c31d43c9623330c46e2

COPY --from=build --chown=65532:65532 /rootfs/ /

ENTRYPOINT ["/tailscaled", "--tun", "userspace-networking", "--statedir", "/state/"]
