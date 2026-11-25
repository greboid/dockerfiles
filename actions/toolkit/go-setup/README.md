# go-setup

Pure-Go replacement for `actions/setup-go` on [Forgejo Runner](https://forgejo.org/docs/latest/user/actions/):
a small static Go binary in an Alpine image that installs a Go toolchain from go.dev with sha256
verification — no Node.js, no `actions/node20` runtime. It talks to the runner through the
standard `GITHUB_*` files (`-flag` args from the action inputs, `$GITHUB_OUTPUT`, `$GITHUB_ENV`, `$GITHUB_PATH`).

## Usage

```yaml
jobs:
  build:
    runs-on: docker # a label backed by a docker:// image
    steps:
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@v1
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/go-setup@v1
        with:
          go-version-file: go.mod
      - run: go build ./...
```

(The `https://` prefix form is unambiguous; bare `owner/repo@ref` also works when the instance
`DEFAULT_ACTIONS_URL` is set to it.)

The image is built from this repository by the images workflow and pushed to the instance
registry; the build needs no network beyond the base images. Running the action needs network
access to go.dev.

## Behaviour

Versions resolve like the original: exact (`1.23.4`), wildcards (`1.23`, `1.23.x`, `1.x`,
`stable`, `latest` — both meaning the newest stable release), or a version file (`go.mod` with
`toolchain`/`go` directives). With no input at all it
falls back to a `go.mod` found in the workspace. Archives are sha256-verified against go.dev
metadata.

Toolchains install under `.actions/go` inside the workspace volume (Forgejo's runner has no shared
tool-cache volume) and persist for the rest of the job; re-running the action reuses them. `PATH`
and `GOROOT` are exported for subsequent steps.

## Caching

When the runner advertises a cache server (`ACTIONS_CACHE_URL` — the same protocol
`actions/cache` speaks, which Forgejo Runner serves itself), the downloaded archive is stored
there under the key `go-setup-<filename>` with the archive's sha256 as the entry version, and
later jobs restore it instead of hitting go.dev. Restored archives are still sha256-verified, so
a corrupted entry can only cost a re-download, never a broken install. All cache operations are
best effort: lookups, downloads and saves that fail log a warning and fall back to go.dev.

Set `cache: false` to opt out. Note the cache only helps if the runner's cache is configured
sanely: a host job containers can reach (`cache.host` — with `container.network: host`, publish
the proxy port and use `127.0.0.1`), fixed `port`/`proxy_port`, and a `cache.dir` on persistent
storage.

## Development

```sh
make check   # go test + go vet (from the repository root)
make images  # build the action image locally, as the images workflow does (from the repository root)
```
