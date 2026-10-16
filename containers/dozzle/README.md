Container image `git.mouse-lake.ts.net/containers/dozzle` — built with apko
from the `packages` organisation's Alpine repository: the image installs the
org's `dozzle` package at an exact pinned version, so the binary,
license notices and version bumps live in `packages/dozzle`
(upstream: https://github.com/amir20/dozzle).

- Tags: `:YYYY.M.D-alpineX.Y` and `:latest`, pushed on every `master` build
- Builds run on push to `master` via the shared workflow in `containers/build-common`
- Package version pins in `apko.yaml` arrive as depbot PRs (`.depbot.yaml`);
  the app version itself is tracked by depbot in the `packages` org's dozzle repo

Runs as nonroot (`65532`) and serves the web UI on port 8080. Mount the host
Docker socket at `/var/run/docker.sock` and give the container access to it
(e.g. `--group-add` for the socket's group on the host) so dozzle can read
container logs.

## Local build

The repository key is committed under `keys/`, so a plain apko build works:

```sh
apko build apko.yaml dozzle:local image.tar
```
