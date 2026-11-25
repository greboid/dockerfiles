# go-test

Runs `go test` on the Go module in the workspace. No `go-setup` needed.

## Inputs

| Input      | Default | Description                              |
| ---------- | ------- | ---------------------------------------- |
| `packages` | `./...` | Packages to test (space separated)       |
| `flags`    |         | Extra `go test` flags, e.g. `-race`      |
| `debug`    | `false` | Enable debug logging                     |

## Outputs

None.

## Example

```yaml
- uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@master
- uses: https://git.mouse-lake.ts.net/actions/toolkit/go-test@master
  with:
    flags: -race -count=1
```
