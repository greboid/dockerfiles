# golang

Container image `git.mouse-lake.ts.net/containers/golang` — built with apko
from the `packages` organisation's Alpine repository: the image installs the
org's `go` package (the source-built toolchain at the legacy `/usr/local/go`
path, with a writable `/go` workspace and the repo gitconfig) plus
`go-licenses-builder` (the license-notices collector), so image consumers and
melange package builds use the exact same toolchain.

- Tags: `:YYYY.M.D-alpineX.Y` and `:latest`, pushed on every `master` build
  (and mirrored to ghcr.io)
- Builds run on push to `master` via the shared workflow in `containers/build-common`
- Package version pins in `apko.yaml` arrive as depbot PRs (`.depbot.yaml`);
  the toolchain itself is tracked by depbot in the `packages` org's go repo

## Local build

The repository key is committed under `keys/`, so a plain apko build works:

```sh
apko build apko.yaml golang:local image.tar
```
