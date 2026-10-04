Container image `git.mouse-lake.ts.net/containers/forgejo-runner` — built with apko
from the `packages` organisation's Alpine repository: the image installs the
org's `forgejo-runner` package at an exact pinned version plus `buildah`, `git`, `fuse-overlayfs` and `gnupg`; `CONTAINERS_STORAGE_CONF` points at the `fuse-overlay-storage-conf`
package's storage.conf, so the binary,
license notices and version bumps live in `packages/forgejo-runner`
(upstream: https://code.forgejo.org/forgejo/runner).

- Tags: `:YYYY.M.D-alpineX.Y` and `:latest`, pushed on every `master` build
- Builds run on push to `master` via the shared workflow in `containers/build-common`
- Package version pins in `apko.yaml` arrive as depbot PRs (`.depbot.yaml`);
  the app version itself is tracked by depbot in the `packages` org's forgejo-runner repo

## Local build

The repository key is committed under `keys/`, so a plain apko build works:

```sh
apko build apko.yaml forgejo-runner:local image.tar
```
