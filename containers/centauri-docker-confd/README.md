Container image `git.mouse-lake.ts.net/containers/centauri-docker-confd` — built with apko
from the `packages` organisation's Alpine repository: the image installs the
org's `centauri-docker-confd` package at an exact pinned version, so the binary,
license notices and version bumps live in `packages/centauri-docker-confd`
(upstream: https://github.com/csmith/centauri-docker-confd).

- Tags: `:YYYY.M.D-alpineX.Y` and `:latest`, pushed on every `master` build
- Builds run on push to `master` via the shared workflow in `containers/build-common`
- Package version pins in `apko.yaml` arrive as depbot PRs (`.depbot.yaml`);
  the app version itself is tracked by depbot in the `packages` org's centauri-docker-confd repo

## Local build

The repository key is committed under `keys/`, so a plain apko build works:

```sh
apko build apko.yaml centauri-docker-confd:local image.tar
```
