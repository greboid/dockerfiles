# docker-build

Build-only sibling of `docker-push` on [Forgejo Runner](https://forgejo.org/docs/latest/user/actions/):
a docker action — a small Go binary in its own container, built from this repository by the runner —
that builds an OCI image with `buildah` (baked into the action image; no Node.js,
no `actions/node20` runtime), with the `platforms` input selecting the target platforms. When tags
were given, a single-platform result is pushed into the runner's engine over
`DOCKER_HOST` so later steps can use it; an untagged build is only reported via the `imageid`
output. A multi-platform build only validates that every platform builds: nothing is loaded into
the engine, even with tags. Nothing is pushed to a registry and no registry credentials are read.

## Usage

```yaml
jobs:
  build:
    runs-on: alpine # a label backed by a docker:// image (alpine works)
    steps:
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@v1
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/docker-build@v1
        with:
          tags: myapp:local
      - run: docker run --rm myapp:local myapp --version
```

(The `https://` prefix form is unambiguous; bare `owner/repo@ref` also works when the instance
`DEFAULT_ACTIONS_URL` is set to it.)

The workspace is the build context root of reference; the image is built inside the action
container, which the runner gives access to the same workspace volume. Building unprivileged needs
a tmpfs big enough for pulled layers to stage in — when the action container's root is overlayfs
the usual case), the build may need the runner job to provide one (see `docker-push`'s notes; the
same `--shm-size=1g` hint applies).

## Inputs

| input        | description                                                                                                                 |
| ------------ | --------------------------------------------------------------------------------------------------------------------------- |
| `tags`       | Full image references to tag the image with, separated by commas or whitespace (optional; untagged builds are only addressable by the `imageid` output) |
| `context`    | Build context directory, relative to the workspace (default `.`)                                                            |
| `file`       | Containerfile/Dockerfile path, relative to the context (default: Containerfile if present, else Dockerfile)                  |
| `target`     | Build target stage                                                                                                          |
| `platforms`  | Target platforms, separated by commas or whitespace (e.g. `linux/amd64,linux/arm64`); more than one builds each platform but loads nothing into the runner engine, whose docker-daemon transport takes only single images |
| `build-args` | Build-time variables, one `KEY=VALUE` per line (values may contain spaces)                                                  |
| `no-cache`   | Do not use the docker build cache                                                                                            |
| `debug`      | Enable debug logging                                                                                                        |

## Outputs

| output    | description                                                     |
| --------- | --------------------------------------------------------------- |
| `imageid` | ID of the built image (with multiple platforms, of the last platform build) |
