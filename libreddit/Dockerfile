# Generated from https://github.com/greboid/dockerfiles/blob/master/libreddit/Containerfile.gotpl
# BOM: {"github:libreddit/libreddit":"v3.0.0","image:base":"5c298191cd12c92b8b3df64ab907f33118d5ddffa99f0c31d43c9623330c46e2","image:rust":"8fec0eb45e1f7ef5004605a6d27537dfd2983d0a6ad9166130a737284b42fedb"}

FROM reg.g5d.dev/rust@sha256:8fec0eb45e1f7ef5004605a6d27537dfd2983d0a6ad9166130a737284b42fedb as server

ARG TAG="v3.0.0"

RUN set -eux; \
apk add git; \
git clone --depth=1 -b $TAG --single-branch https://github.com/libreddit/libreddit /src; \
cd /src

WORKDIR /src

RUN set -eux; \
apk add musl-dev gcc openssl-dev; \
cargo build --release

FROM reg.g5d.dev/base@sha256:5c298191cd12c92b8b3df64ab907f33118d5ddffa99f0c31d43c9623330c46e2
COPY --from=server /src/target/release/libreddit /
ENTRYPOINT ["/libreddit"]
