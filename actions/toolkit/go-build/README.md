# go-build

Builds a static binary from the Go module in the workspace. No `go-setup` needed.

## Inputs

| Input      | Default                              | Description                              |
| ---------- | ------------------------------------ | ---------------------------------------- |
| `cgo`      | `0`                                  | Set to `1` to enable cgo                  |
| `ldflags`  | `-w -extldflags "-static" -buildid=.` | Flags passed to `-ldflags`               |
| `packages` | `.`                                  | Packages to build (space separated)      |
| `output`   | module name                          | Output path or binary name               |
| `debug`    | `false`                              | Enable debug logging                     |

## Outputs

| Output   | Description                       |
| -------- | --------------------------------- |
| `binary` | Absolute path of the built binary |

## Example

```yaml
- uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@master
- uses: https://git.mouse-lake.ts.net/actions/toolkit/go-build@master
  with:
    packages: ./cmd/myapp
    output: myapp
```
