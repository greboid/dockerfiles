# rust

Container image `git.mouse-lake.ts.net/containers/rust` — built with apko
from the `packages` organisation's Alpine repository: the image installs the
org's `rust-toolchain` package (pinned minimal-profile rustup with a musl
host toolchain) at an exact pinned version, plus the standard C build tools.

- Tags: `:YYYY.M.D-alpineX.Y` and `:latest`, pushed on every `master` build
- Builds run on push to `master` via the shared workflow in `containers/build-common`
- Package version pins in `apko.yaml` arrive as depbot PRs (`.depbot.yaml`);
  the toolchain itself is tracked by depbot in the `packages` org's rust-toolchain repo

## Local build

The repository key is committed under `keys/`, so a plain apko build works:

```sh
apko build apko.yaml rust:local image.tar
```
