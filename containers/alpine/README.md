# alpine

Container image `git.mouse-lake.ts.net/containers/alpine` — an Alpine 3.24
rootfs built with apko instead of the upstream minirootfs tarball: the same
floor packages every fleet image relies on (busybox, musl, CA certificates,
timezones, `alpine-baselayout`) plus the `packages` organisation's
`os-release`, pinned exactly and installed straight from the org's Alpine
repository.

- Tags: `:YYYY.M.D-alpineX.Y` and `:latest`, pushed on every `master` build
  (and mirrored to ghcr.io)
- Builds run on push to `master` via the shared workflow in `containers/build-common`
- Package version pins in `apko.yaml` arrive as depbot PRs (`.depbot.yaml`)

## Local build

The repository key is committed under `keys/`, so a plain apko build works:

```sh
apko build apko.yaml alpine:local image.tar
```
