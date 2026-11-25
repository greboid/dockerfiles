# go-setup

Installs a Go toolchain and adds it to `PATH` for later steps. With no inputs, the version is read from `go.mod`.

## Inputs

| Input             | Default            | Description                                                     |
| ----------------- | ------------------ | --------------------------------------------------------------- |
| `go-version`      |                    | Version to install: `1.23.4`, `1.23`, `1.x`, `stable` or `latest` |
| `go-version-file` | `go.mod`           | File to read the version from                                   |
| `install-dir`     | `.actions/go`      | Directory to install into                                       |
| `force`           | `false`            | Re-download even if already installed                           |
| `cache`           | `true`             | Cache the download in the runner cache                          |
| `debug`           | `false`            | Enable debug logging                                            |

## Outputs

| Output       | Description                 |
| ------------ | --------------------------- |
| `go-version` | The installed Go version    |

## Example

```yaml
- uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@master
- uses: https://git.mouse-lake.ts.net/actions/toolkit/go-setup@master
  with:
    go-version: stable
- run: go build ./...
```
