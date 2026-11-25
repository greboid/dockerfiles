# package-latest

Finds the latest version of a package in this Forgejo instance's package registry, and the next version to release (patch number incremented).

## Inputs

| Input        | Default          | Description                                                   |
| ------------ | ---------------- | ------------------------------------------------------------- |
| `server-url` | this instance    | Forgejo URL to query                                          |
| `owner`      | repository owner | User or org that owns the package                             |
| `type`       | all types        | Package type, e.g. `container` or `generic`                   |
| `package`    |                  | Package name to match                                         |
| `prefix`     |                  | Only consider versions starting with this, e.g. `1.20261007.` |
| `token`      |                  | Access token, for private packages                            |
| `debug`      | `false`          | Enable debug logging                                          |

## Outputs

| Output    | Description                                                                            |
| --------- | -------------------------------------------------------------------------------------- |
| `version` | Latest matching version, empty if none                                                 |
| `next`    | Next version: `1.2.3` → `1.2.4`; with no match, `<prefix>0` (or `0.0.0` without a prefix) |

## Example

```yaml
- uses: https://git.mouse-lake.ts.net/actions/toolkit/package-latest@master
  id: pkg
  with:
    type: container
    package: myorg/myimage
    prefix: 1.20261007.
- run: echo "Next version is ${{ steps.pkg.outputs.next }}"
```
