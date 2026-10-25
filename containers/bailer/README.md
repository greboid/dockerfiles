Container image `git.mouse-lake.ts.net/containers/bailer` — built with apko
from the `packages` organisation's Alpine repository: the image installs the
org's `bailer` and `restic` packages at exact pinned versions, so the
binaries, license notices and version bumps live in `packages/bailer`
(upstream: https://git.mouse-lake.ts.net/greboid/bailer) and
`packages/restic` (upstream: https://github.com/restic/restic). `restic` is
included so bailer's restic backend works out of the box.

- Tags: `:YYYY.M.D-alpineX.Y` and `:latest`, pushed on every `master` build
- Builds run on push to `master` via the shared workflow in `containers/build-common`
- Package version pins in `apko.yaml` arrive as depbot PRs (`.depbot.yaml`);
  the app versions themselves are tracked by depbot in the `packages` org's
  bailer and restic repos

## Local build

The repository key is committed under `keys/`, so a plain apko build works:

```sh
apko build apko.yaml bailer:local image.tar
```
