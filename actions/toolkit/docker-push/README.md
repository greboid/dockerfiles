# docker-push

Replacement for `docker/build-push-action` on [Forgejo Runner](https://forgejo.org/docs/latest/user/actions/):
a docker action — a small Go binary in its own container, built from this repository by the runner —
that builds OCI images with `buildah` (baked into the action image; no Node.js, no
`actions/node20` runtime) and pushes each tag straight to its registry, optionally mirroring to
GHCR. One platform is built by default; the `platforms` input names several and produces a
multi-platform manifest list per tag. With no `tags` input, tags are derived from the workflow ref
(the rules of greboid/docker-tags-action) as `<registry>/<owner>/<name>:<version>` against the
`registries` input, defaulting to the Forgejo instance of this workflow.

## Usage

```yaml
jobs:
  build:
    runs-on: alpine # a label backed by a docker:// image (alpine works)
    steps:
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@v1
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/docker-login@v1
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/docker-push@v1
        if: github.ref == 'refs/heads/master'
```

With `docker-login` before it (or no inputs at all, which logs in to this instance's registry with
the job token), pushing to the instance registry needs no further configuration: a `master` build
pushes `<instance-registry>/<owner>/<repo>:dev`, a `v1.2.3` tag pushes
`:latest` (when it is the newest stable major), `:1.2.3`, `:1.2` and `:1`.

(The `https://` prefix form is unambiguous; bare `owner/repo@ref` also works when the instance
`DEFAULT_ACTIONS_URL` is set to it.)

### Tag derivation rules

| workflow ref                    | derived tags                                  |
| ------------------------------- | --------------------------------------------- |
| `refs/heads/master` or `main`   | `dev`                                         |
| `refs/tags/v1.2.3`              | `1.2.3`, `1.2`, `1`, plus `latest` when 1 is the newest stable tag major in the repository |
| anything else                   | none (the step fails; set `tags` explicitly)  |

The `latest` rule lists the repository's tags through the instance API using the `token` input
(`GITHUB_TOKEN` by default); pre-release suffixes are dropped from the generated tags.

## Inputs

| input             | description                                                                                                          |
| ----------------- | -------------------------------------------------------------------------------------------------------------------- |
| `tags`            | Full image references to tag and push, separated by commas or whitespace; derived from the workflow ref when empty   |
| `registries`      | Comma separated registries for derived tags; defaults to the registry of the Forgejo instance this workflow runs on  |
| `repository`      | Image name for derived tags; defaults to owner/repository-name                                                       |
| `token`           | Token used to list repository tags for the `latest` rule; defaults to `GITHUB_TOKEN`                                 |
| `context`         | Build context directory, relative to the workspace (default `.`)                                                     |
| `file`            | Containerfile/Dockerfile path, relative to the context (default: Containerfile if present, else Dockerfile)          |
| `target`          | Build target stage                                                                                                   |
| `platforms`       | Target platforms, separated by commas or whitespace (e.g. `linux/amd64,linux/arm64`); more than one builds each platform into a shared manifest list pushed with every tag |
| `build-args`      | Build-time variables, one `KEY=VALUE` per line (values may contain spaces)                                           |
| `push`            | Push after building (default `true`)                                                                                  |
| `no-cache`        | Do not use the docker build cache                                                                                     |
| `registry`        | Registry for the explicit credentials below (optional; otherwise the docker config from `docker-login` is used)      |
| `username`        | Registry username (optional; same fallback)                                                                           |
| `password`        | Registry password (optional; same fallback)                                                                           |
| `ghcr`            | Mirror the built image to the GitHub Container Registry (`ghcr.io/owner/repo`) in addition to the tags                |
| `ghcr-repository` | GHCR repository to mirror to, owner/name (defaults to `GITHUB_REPOSITORY`, lower-cased)                               |
| `ghcr-token`      | GitHub PAT with `write:packages` to push the mirror; defaults to ghcr.io credentials from `docker-login`              |
| `ghcr-username`   | Username for the ghcr.io login (defaults to `GITHUB_ACTOR`)                                                            |
| `debug`           | Enable debug logging                                                                                                  |

## Outputs

| output    | description                                                                   |
| --------- | ----------------------------------------------------------------------------- |
| `imageid` | ID of the built image (with multiple platforms, of the last platform build)    |
| `digest`  | Digest of the pushed manifest (of the manifest list for multi-platform pushes) |
| `version` | The pushed semver for tag refs, the commit SHA on master/main, `unknown` otherwise |
| `tags`    | The image references pushed, comma separated                                    |

## Notes

- The build runs unprivileged inside the action container: the storage driver is probed
  (fuse-overlayfs, falling back to vfs), isolation is chroot. When the container root is overlayfs,
  TMPDIR must be a tmpfs large enough for pulled layers to stage in — if pushes of larger images
  fail with "no space left on device", give the job container a bigger tmpfs
  (`container.options: --shm-size=1g` on the runner job).
- The `ghcr` mirror pushes each distinct tag suffix once, regardless of how many registries the
  tags list names; in the multi-platform flow it pushes the manifest list itself.
- Multi-platform builds run one `buildah bud --platform <p> --manifest <list>` per platform and
  push each tag with `buildah manifest push --all`, so the registry serves a manifest list and
  `docker pull` selects the right platform. When the Containerfile pins its build stage to
  `--platform=${BUILDPLATFORM}` and cross-compiles (as with `CGO_ENABLED=0` Go builds), the extra
  platforms need no arm64 runner and no emulation; final stages must not `RUN` anything, since
  only the build stage executes on the host platform.
