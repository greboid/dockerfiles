Container image `git.mouse-lake.ts.net/containers/beszel` — built with apko
from the `packages` organisation's Alpine repository: the image installs the
org's `beszel` package at an exact pinned version, so the binary,
license notices and version bumps live in `packages/beszel`
(upstream: https://github.com/henrygd/beszel).

This is the hub (web interface and coordinator); the agent image is
`containers/beszel-agent`. The single `beszel` package ships both binaries.

- Tags: `:YYYY.M.D-alpineX.Y` and `:latest`, pushed on every `master` build
- Builds run on push to `master` via the shared workflow in `containers/build-common`
- Package version pins in `apko.yaml` arrive as depbot PRs (`.depbot.yaml`);
  the app version itself is tracked by depbot in the `packages` org's beszel repo

## Running

The hub listens on 8090 and stores its database and SSH host key in
`/beszel_data` (owned by the nonroot user 65532). Mount a volume there:

```sh
docker run -d -p 8090:8090 -v beszel_data:/beszel_data \
  git.mouse-lake.ts.net/containers/beszel
```

Default arguments can be overridden via `cmd`; see `beszel serve --help`.
Configure the hub via env vars (`BESZEL_*`).

## Local build

The repository key is committed under `keys/`, so a plain apko build works:

```sh
apko build apko.yaml beszel:local image.tar
```
