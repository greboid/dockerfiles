# git-checkout

Checks out a git repository into the workspace. Defaults to the workflow's repository and ref.

## Inputs

| Input         | Default             | Description                                      |
| ------------- | ------------------- | ------------------------------------------------ |
| `repository`  | this repository     | Repository to check out (`owner/name`)           |
| `ref`         | the workflow ref    | Branch, tag or SHA to check out                  |
| `token`       | `GITHUB_TOKEN`      | Access token                                     |
| `username`    | `GITHUB_ACTOR`      | Username for the token                           |
| `path`        | `.`                 | Directory to check out into                      |
| `fetch-depth` | `1`                 | Number of commits to fetch; `0` fetches all history and tags |
| `submodules`  | `false`             | `false`, `true` or `recursive`                   |
| `debug`       | `false`             | Enable debug logging                             |

## Outputs

| Output   | Description                    |
| -------- | ------------------------------ |
| `ref`    | The ref that was checked out   |
| `commit` | The commit SHA checked out     |

## Example

```yaml
- uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@master
  with:
    fetch-depth: 0
```
