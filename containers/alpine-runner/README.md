# alpine-runner

Container image `git.mouse-lake.ts.net/containers/alpine-runner` — the CI
runner image Forgejo's `runs-on: alpine` label maps to. Built with apko from
the `packages` organisation's Alpine repository: it installs the org's
`alpine-runner` package (the overlay-over-fuse storage configuration and
legacy root passwd entries), `buildah` (org package, fuse-overlayfs mount
program), plus git, curl, wget, openssl and gnupg for the toolkit actions.

- Tags: `:YYYY.M.D-alpineX.Y` and `:latest`, pushed on every `master` build
- Builds run on push to `master` via the shared workflow in `containers/build-common`
- Package version pins in `apko.yaml` arrive as depbot PRs (`.depbot.yaml`)

## Local build

The repository key is committed under `keys/`, so a plain apko build works:

```sh
apko build apko.yaml alpine-runner:local image.tar
```
