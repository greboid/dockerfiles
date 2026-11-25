# git-try-push

Pushes the current branch (and optional tags), retrying on failure. Between attempts it fetches and rebases onto the remote branch, so concurrent jobs pushing to the same branch don't fail. Run `git-checkout` first.

## Inputs

| Input          | Default          | Description                                             |
| -------------- | ---------------- | ------------------------------------------------------- |
| `branch`       | current branch   | Branch to push to                                       |
| `tags`         |                  | Tags to push as well (whitespace separated)             |
| `remote`       | `origin`         | Remote to push to                                       |
| `retries`      | `10`             | Number of retries                                       |
| `initial_wait` | `2`              | Seconds to wait before the first retry                  |
| `wait_factor`  | `2`              | Multiplier applied to the wait after each retry         |
| `wait_max`     | `60`             | Maximum wait in seconds                                 |
| `rebase`       | `true`           | Fetch and rebase between attempts                       |
| `push_flags`   |                  | Extra `git push` flags, e.g. `--force-with-lease`       |
| `token`        | `GITHUB_TOKEN`   | Access token                                            |
| `audience`     |                  | Forgejo Authorized Integration audience to push with instead of a token |
| `username`     | `GITHUB_ACTOR`   | Username for the token                                  |
| `debug`        | `false`          | Enable debug logging                                    |

## Outputs

| Output | Description         |
| ------ | ------------------- |
| `head` | The SHA pushed      |

## Example

```yaml
- uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@master
- run: ./bump-version.sh
- uses: https://git.mouse-lake.ts.net/actions/toolkit/git-commit@master
  with:
    message: Bump version
    push: false
- uses: https://git.mouse-lake.ts.net/actions/toolkit/git-try-push@master
```
