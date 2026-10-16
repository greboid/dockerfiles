Container image `git.mouse-lake.ts.net/containers/beszel-agent` — built with
apko from the `packages` organisation's Alpine repository: the image installs
the org's `beszel` package at an exact pinned version, so the binary,
license notices and version bumps live in `packages/beszel`
(upstream: https://github.com/henrygd/beszel).

This is the agent (metrics collector for a monitored host); the hub image is
`containers/beszel`. The single `beszel` package ships both binaries.

- Tags: `:YYYY.M.D-alpineX.Y` and `:latest`, pushed on every `master` build
- Builds run on push to `master` via the shared workflow in `containers/build-common`
- Package version pins in `apko.yaml` arrive as depbot PRs (`.depbot.yaml`);
  the app version itself is tracked by depbot in the `packages` org's beszel repo

## Running

Runs as nonroot (65532). Configure via env vars at runtime — `KEY` (or
`KEY_FILE`) must hold the agent's public key from the hub; `PORT`/`LISTEN`
override the default listen port (45876); `HUB_URL`/`TOKEN` for
WebSocket-token mode. Mount `/var/run/docker.sock` for Docker container
stats:

```sh
docker run -d -p 45876:45876 \
  -e KEY=<public key from the hub> \
  -v /var/run/docker.sock:/var/run/docker.sock \
  git.mouse-lake.ts.net/containers/beszel-agent
```

## Local build

The repository key is committed under `keys/`, so a plain apko build works:

```sh
apko build apko.yaml beszel-agent:local image.tar
```
