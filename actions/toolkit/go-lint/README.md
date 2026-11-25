# go-lint

gofmt check for [Forgejo Runner](https://forgejo.org/docs/latest/user/actions/): a docker
action — a small Go binary in a Go toolchain container, built from this repository by the
runner — that fails the step when any of the workspace module's own `.go` files are not
gofmt-formatted, listing the offenders. No Node.js, no `actions/node20` runtime; the gofmt is
the one shipping with the action image's toolchain.

## Usage

```yaml
jobs:
  check:
    runs-on: alpine # a label backed by a docker:// image (alpine works)
    steps:
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@v1
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/go-lint@v1
```

(The `https://` prefix form is unambiguous; bare `owner/repo@ref` also works when the instance
`DEFAULT_ACTIONS_URL` is set to it.)

## Inputs

| input   | description          |
| ------- | -------------------- |
| `debug` | Enable debug logging |

## Outputs

None.

## Notes

- Only the module's own files are checked: the trees the go tool ignores for `./...` patterns —
  dot- and underscore-prefixed directories (which also keeps installs under `.actions` out) and
  `vendor` directories — are pruned, so the check flags first-party code only.
- The workspace must contain a `go.mod`.
- Fix offenders with `gofmt -w` (or `go fmt ./...`).
