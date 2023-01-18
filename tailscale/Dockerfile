# Generated from https://github.com/greboid/dockerfiles/blob/master/tailscale/Dockerfile.gotpl
# BOM: {"github:tailscale/tailscale":"v1.34.2","image:base":"ffd04efedafc54f5392c35a1045604c12a4329e5780407d5fe08dfe495d411b1","image:golang":"62336f81c6c2c3ecdf35de9cf62c7ad3e84cebad54841ad004c734da9ca56bbe"}

FROM reg.g5d.dev/golang@sha256:62336f81c6c2c3ecdf35de9cf62c7ad3e84cebad54841ad004c734da9ca56bbe AS build

ARG TAG="v1.34.2"

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

FROM reg.g5d.dev/base@sha256:ffd04efedafc54f5392c35a1045604c12a4329e5780407d5fe08dfe495d411b1

COPY --from=build --chown=65532:65532 /rootfs/ /

ENTRYPOINT ["/tailscaled", "--tun", "userspace-networking", "--statedir", "/state/"]
