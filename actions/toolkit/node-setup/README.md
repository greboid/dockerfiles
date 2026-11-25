# node-setup

Pure-Go replacement for `actions/setup-node` on [Forgejo Runner](https://forgejo.org/docs/latest/user/actions/):
a small pure-Go program shipped as a docker image (`runs: {using: docker,
image: docker://...}`) that installs a Node.js distribution from nodejs.org with sha256
verification — no Node.js runtime to bootstrap, no `actions/node20`, and no Go toolchain on the
runner. It talks to the runner through the standard
`GITHUB_*` files (`-flag` args from the action inputs, `$GITHUB_OUTPUT`, `$GITHUB_ENV`, `$GITHUB_PATH`).

## Usage

```yaml
jobs:
  build:
    runs-on: docker # a label backed by a docker:// image
    steps:
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@v1
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/node-setup@v1
        with:
          node-version: lts/*
      - run: pnpm install --frozen-lockfile
```

(The `https://` prefix form is unambiguous; bare `owner/repo@ref` also works when the instance
`DEFAULT_ACTIONS_URL` is set to it.)

The image — a multi-stage build that compiles the static Go binary and embeds
it in a small Alpine image — is built by the images workflow and pulled by the runner, which
runs it with the workspace mounted. Running it needs network
access to nodejs.org.

## Behaviour

Versions resolve like the original: exact (`18.19.0`), wildcards (`18`, `18.19`, `lts/*`,
`lts/Jod`, `latest`), or a version file (`.nvmrc`, `package.json` `engines.node`; a
`package.json` without `engines.node` falls back to `latest`). With no input
at all it falls back to a `.nvmrc` or `package.json` found in the working directory — the
workspace by default, or `working-dir` for monorepos where the Node project lives in a
subdirectory. Relative `node-version-file` paths resolve against the working directory too.
Archives are sha256-verified against `SHASUMS256.txt`. The official builds are linked against
glibc; the action detects the libc of the image it runs in — its own Alpine image — and downloads
the musl builds from [unofficial-builds.nodejs.org](https://unofficial-builds.nodejs.org) instead,
matching the Alpine-style images the installed toolchain is used on. Those builds
load libstdc++/libgcc dynamically, so when the image does not provide them the action bundles
Alpine's `libstdc++`/`libgcc` packages (checksum-verified against the repository's `APKINDEX`)
into the toolchain and exports `LD_LIBRARY_PATH`.

On top of Node the action sets up a package manager, selected with `package-manager`: `pnpm`
(the default), `yarn` or `npm`. pnpm and yarn are installed into the toolchain with the Node
distribution's own npm — pinned to the version from a corepack-style `packageManager` field in
`package.json` (e.g. `pnpm@9.15.0+sha512…`) when it names the same manager, `latest` otherwise;
`npm` just uses the copy bundled with Node. The install lands inside the toolchain directory, so
it is reused whenever the toolchain is.

Distributions install under `.actions/node` inside the workspace volume (Forgejo's runner has no
shared tool-cache volume) and persist for the rest of the job; re-running the action reuses them.
`PATH` is exported for subsequent steps. There is no cross-job cache — if you need one, mount a
volume at `.actions` or set `install-dir`. `registry-url` writes a project `.npmrc` with the token
(default `GITHUB_TOKEN`) in the working directory, e.g. for a Forgejo npm registry.

## Development

```sh
make check   # go test + go vet (from the repository root)
make images  # build the action image locally, as the images workflow does (from the repository root)
```
