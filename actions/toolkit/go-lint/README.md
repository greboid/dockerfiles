# go-lint

Fails if any Go files in the workspace module are not `gofmt` formatted, and lists them. No `go-setup` needed.

## Inputs

| Input   | Default | Description          |
| ------- | ------- | -------------------- |
| `debug` | `false` | Enable debug logging |

## Outputs

None.

## Example

```yaml
- uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@master
- uses: https://git.mouse-lake.ts.net/actions/toolkit/go-lint@master
```
