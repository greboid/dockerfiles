# go-vet

Runs `go vet` on the Go module in the workspace. No `go-setup` needed.

## Inputs

| Input      | Default | Description                       |
| ---------- | ------- | --------------------------------- |
| `packages` | `./...` | Packages to vet (space separated) |
| `debug`    | `false` | Enable debug logging              |

## Outputs

None.

## Example

```yaml
- uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@master
- uses: https://git.mouse-lake.ts.net/actions/toolkit/go-vet@master
```
