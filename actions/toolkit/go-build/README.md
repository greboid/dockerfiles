# go-build

`go build` step for [Forgejo Runner](https://forgejo.org/docs/latest/user/actions/): a docker
action — a small Go binary in a Go toolchain container, built from this repository by the
runner — that compiles the Go module checked out in the workspace with the static-build flags
shared by the other actions: `-tags netgo,osusergo -a -trimpath`, CGO disabled by default and
ldflags that strip and statically link the result. No Node.js, no `actions/node20` runtime; the
toolchain (plus gofmt and git for VCS stamping) is baked into the action image.

## Usage

```yaml
jobs:
  build:
    runs-on: alpine # a label backed by a docker:// image (alpine works)
    steps:
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@v1
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/go-build@v1
```

(The `https://` prefix form is unambiguous; bare `owner/repo@ref` also works when the instance
`DEFAULT_ACTIONS_URL` is set to it.)

The build runs inside the action container against the workspace volume; the module and build
caches are kept in the workspace (`.actions/gomodcache`, `.actions/gobuildcache`) so repeated
go-* steps reuse them. Enabling cgo requires a C compiler — the default golang image ships one.

## Inputs

| input     | description                                                                                |
| --------- | ------------------------------------------------------------------------------------------ |
| `cgo`     | Enable cgo (sets `CGO_ENABLED=1`); requires a C compiler in the action image (default `0`) |
| `ldflags` | Flags passed to `-ldflags` verbatim (default `-w -extldflags "-static" -buildid=.`)        |
| `packages` | Space-separated package patterns to build (default `.`)                                   |
| `output`  | Output path or binary name, relative to the workspace; defaults to the module name from go.mod |
| `debug`   | Enable debug logging                                                                       |

## Outputs

| output   | description                        |
| -------- | ---------------------------------- |
| `binary` | Absolute path of the built binary  |

## Notes

- The workspace must contain a `go.mod` (the default build target is the main package at the
  workspace root; pass `packages` to build others — with multiple packages `output` must name a
  directory).
- This action needs no `go-setup`: the action image is the toolchain. `go-setup` remains useful
  when *your own steps* need a Go toolchain in the job container, or to pin a different version
  for them.
