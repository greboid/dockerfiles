# go-test

`go test` step for [Forgejo Runner](https://forgejo.org/docs/latest/user/actions/): a docker
action — a small Go binary in a Go toolchain container, built from this repository by the
runner — that runs `go test` over the module checked out in the workspace, `./...` by default,
so the whole module including the tests of internal packages. No Node.js, no
`actions/node20` runtime; the toolchain is baked into the action image.

## Usage

```yaml
jobs:
  test:
    runs-on: alpine # a label backed by a docker:// image (alpine works)
    steps:
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@v1
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/go-test@v1
        with:
          flags: -race -count=1
```

(The `https://` prefix form is unambiguous; bare `owner/repo@ref` also works when the instance
`DEFAULT_ACTIONS_URL` is set to it.)

The module and build caches are kept in the workspace (`.actions/gomodcache`,
`.actions/gobuildcache`) so repeated go-* steps reuse them.

## Inputs

| input      | description                                                        |
| ---------- | ------------------------------------------------------------------ |
| `packages` | Space-separated package patterns to test (default `./...`)          |
| `flags`    | Extra whitespace-separated go test flags, e.g. `-race -count=1`     |
| `debug`    | Enable debug logging                                               |

## Outputs

None.

## Notes

- The workspace must contain a `go.mod` (the module is tested at the workspace root).
- This action needs no `go-setup`: the action image is the toolchain.
