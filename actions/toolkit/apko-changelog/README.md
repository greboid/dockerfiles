# apko-changelog

Compares two apko lock files and produces a changelog message listing the package changes per architecture.

## Inputs

| Input          | Default          | Description                                                     |
| -------------- | ---------------- | --------------------------------------------------------------- |
| `new-lockfile` | `base.lock.json` | Path of the new lock file                                       |
| `old-lockfile` | `old.lock.json`  | Path of the previous lock file                                  |
| `version`      | `$VERSTRING`     | Version string mentioned in the message                         |
| `debug`        | `false`          | Enable debug logging                                            |

## Outputs

| Output    | Description                                  |
| --------- | -------------------------------------------- |
| `changes` | `true` when any package changed              |
| `message` | The changelog message (multi-line)           |

## Example

```yaml
- uses: https://git.mouse-lake.ts.net/actions/toolkit/apko-changelog@master
  id: changelog
  with:
    new-lockfile: base.lock.json
    old-lockfile: old.lock.json
    version: "20260714"
- run: echo "${{ steps.changelog.outputs.message }}"
```

Example message:

```
Updating to 20260714 with the following package changes:
 - x86_64
   - tzdata: 2026d-r0 => 2026e-r0
```
