# Generated from https://github.com/greboid/dockerfiles/blob/master/libreddit/Dockerfile.gotpl
# BOM: {"github:libreddit/libreddit":"v3.0.0","image:base":"4ddff4468810fe9d9dc3285917d2923c41ae9fb6d1fa2b91c3f1d460cb94427a","image:rust":"a1a43069e5c46de401c93a5c385d953d18149c8e427e58ba1cbecef09393a90d"}

FROM reg.g5d.dev/rust@sha256:a1a43069e5c46de401c93a5c385d953d18149c8e427e58ba1cbecef09393a90d as server

ARG TAG="v3.0.0"

RUN set -eux; \
apk add git; \
git clone --depth=1 -b $TAG --single-branch https://github.com/libreddit/libreddit /src; \
cd /src

WORKDIR /src

RUN set -eux; \
apk add musl-dev gcc openssl-dev; \
cargo build --release

FROM reg.g5d.dev/base@sha256:4ddff4468810fe9d9dc3285917d2923c41ae9fb6d1fa2b91c3f1d460cb94427a
COPY --from=server /src/target/release/libreddit /
ENTRYPOINT ["/libreddit"]
