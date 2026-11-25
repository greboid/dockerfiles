# toolkit

The `actions/*` actions for [Forgejo Runner](https://forgejo.org/docs/latest/user/actions/)
in one repository, sharing one Go module. Every action is a docker action in the style of a
small Go program: a `<action>/<action>.go` package holding the logic and a `<action>/cmd/main.go`
entry point that parses the inputs as command line flags (secrets travel through the
environment instead, so they never appear in argv), built by the runner into a container whose
image bundles every CLI the action needs — no Node.js, no `actions/node20` runtime, no
installing tools at run time.

## Layout

| directory       | purpose                                                                        |
| --------------- | ------------------------------------------------------------------------------ |
| `git-checkout/` | repository checkout over git smart HTTP (go-git, vendored)                     |
| `git-commit/`   | commit workspace changes (all or pathspecs), optionally push                   |
| `git-try-push/` | push with retries: fetch + rebase between attempts                             |
| `release/`      | create a Forgejo release and upload its assets (Authorized Integration auth)   |
| `go-setup/`     | install a Go toolchain from go.dev with checksum verification                  |
| `go-build/`     | static `go build` of the workspace module in a Go toolchain container          |
| `go-test/`      | `go test` over the workspace module                                            |
| `go-vet/`       | `go vet` over the workspace module                                             |
| `go-lint/`      | gofmt check over the workspace module's own files                              |
| `apko-changelog/` | changelog message from a diff of two apko lock files                         |
| `package-latest/` | latest version of a registry package, and the next version to release        |
| `node-setup/`   | install a Node.js distribution + pnpm/yarn/npm                                 |
| `docker-login/` | verify registry credentials (buildah login), store config                      |
| `docker-build/` | build with buildah; single-platform results loaded into the engine             |
| `docker-push/`  | single- or multi-platform build + push with buildah, tags derived from the workflow ref when not given |
| `internal/`     | shared Go packages                                                              |

Shared pieces in `internal/`:

- `gha` — the workflow `Context` (resolved once from the `FORGEJO_*`/`GITHUB_*` environment,
  Forgejo's prefix winning), outputs/env/PATH exports, masking, and the slog-based
  `ConfigureLogging` (`-debug` switches on debug records)
- `buildah` — unprivileged buildah preparation (storage driver probe, tmpfs TMPDIR, chroot
  isolation) plus the invocation and image-reference helpers the docker-* actions share
  and the optional caches below
- `gitcmd` — git CLI runner with the token auth chain (GIT_ASKPASS, or an authorized
  integration JWT as an `http.extraheader`), branch and identity lookups
- `imagetags` — image tag derivation from the workflow ref (docker-push)
- `oidc` — ID token fetch from the runner's OIDC endpoint (Forgejo Authorized Integrations)
- `gotool` — go/gofmt discovery and workspace-backed module/build caches for the go-* actions
- `cache` — runner artifact cache client (the ACTIONS_CACHE_URL protocol)
- `dl` — verified HTTP downloads
- `tarx` — tar.gz/tar.xz extraction
- `ver` — version list/pattern resolution

## Caches

Both caches are opt-in through environment variables set in the runner's `runner.envs`, with
the directories mounted through `container.options` (and listed in `container.valid_volumes`):

- `BUILDAH_CACHE_DIR` — docker-build and docker-push keep images and layers in
  `$BUILDAH_CACHE_DIR/storage`, building with `--layers --pull=newer`. Needs the overlay storage
  driver (fuse-overlayfs); ignored on vfs. Nothing prunes the store.
- `GOMODCACHE` — used by the go-* actions in place of the workspace default, and mounted over
  `/go/pkg/mod` in the `RUN` instructions of docker-build and docker-push.

## Usage

Reference an action by its subdirectory path:

```yaml
steps:
  - uses: actions/toolkit/git-checkout@master
  - uses: actions/toolkit/go-setup@master
  - uses: actions/toolkit/go-test@master
```

The same with an unambiguous fully qualified URL:

```yaml
  - uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@master
```

## How the runner runs these

Every action declares `runs: using: docker` with `image: docker://git.mouse-lake.ts.net/actions/toolkit/<action>:dev` — the runner
pulls the image from the instance registry and runs it with the workspace mounted. The images
are built and pushed on every push to `master` by the images workflow
(`.forgejo/workflows/images.yml`), which builds each root `Dockerfile.<action>` with the whole
repository as the build context. The Dockerfiles live at the repository root — a `Dockerfile`
inside the action directory would make the context just that subdirectory, hiding `internal/`
and the root `go.mod`. The Dockerfiles are the usual two-stage
`golang` builds compiling `./<action>/cmd` from the root module, onto a runtime image that
bundles the action's tools: `buildah`/`netavark`/`fuse-overlayfs` for the docker-* actions,
`git` for the git-* ones, and the `golang` image itself for the go-* ones.

Images track `master` under the `dev` tag, so a workflow pinned to an action's commit SHA still
runs that ref's latest `master` image. To pin the image too, push a matching image tag and
reference it in `action.yml`.

Inputs reach the action as `-flag=${{ inputs.x }}` arguments declared in `action.yml`; secret
inputs (`token`, `password`, ...) are passed through `runs.env` instead so they never appear in
the container's command line, and are masked before use.

The `vendor/` tree is committed, so docker builds need no Go module proxy.

## Versioning

Everything moves together on `master`. To pin, reference a commit SHA; to
version a single action, push a namespaced tag such as `go-build-v1` and
reference `actions/toolkit/go-build@go-build-v1`.

## Development

```sh
make check   # go test + go vet + action.yml validation
make images  # build every docker action image locally, as the images workflow does
```

`./bootstrap.sh` builds and pushes the images the images workflow would publish —
same refs, `git.mouse-lake.ts.net/actions/toolkit/<action>:dev` — from a machine
with docker or buildah. Use it to bootstrap the registry or to publish without
waiting for CI; credentials come from `REGISTRY`/`USERNAME`/`PASSWORD` env.

Each action directory has its own README describing inputs and outputs.
History: the per-action repositories under `actions/` were merged here with
`git subtree`, so `git log --follow <action>/<action>.go` reaches the original
commits.
