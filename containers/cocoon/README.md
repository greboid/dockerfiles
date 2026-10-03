Container image `git.mouse-lake.ts.net/containers/cocoon` — built with apko
from the `packages` organisation's Alpine repository: the image installs the
org's `cocoon` package at an exact pinned version, so the binary,
license notices and version bumps live in `packages/cocoon`
(upstream: https://github.com/haileyok/cocoon).

- Tags: `:YYYY.M.D-alpineX.Y` and `:latest`, pushed on every `master` build
- Builds run on push to `master` via the shared workflow in `containers/build-common`
- Package version pins in `apko.yaml` arrive as depbot PRs (`.depbot.yaml`);
  the app version itself is tracked by depbot in the `packages` org's cocoon repo

## Local build

The repository key is committed under `keys/`, so a plain apko build works:

```sh
apko build apko.yaml cocoon:local image.tar
```
