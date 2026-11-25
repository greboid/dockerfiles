# git-commit

Commits changes in the workspace and optionally pushes them. Does nothing if there are no changes. Run `git-checkout` first.

## Inputs

| Input      | Default                 | Description                                             |
| ---------- | ----------------------- | ------------------------------------------------------- |
| `message`  |                         | Commit message (required)                               |
| `files`    | all changes             | Paths/globs to commit (whitespace separated)            |
| `name`     | `GITHUB_ACTOR`          | Committer name                                          |
| `email`    | noreply address         | Committer email                                         |
| `push`     | `true`                  | Push after committing                                   |
| `branch`   | current branch          | Branch to push to                                       |
| `token`    | `GITHUB_TOKEN`          | Access token for pushing                                |
| `audience` |                         | Forgejo Authorized Integration audience to push with instead of a token |
| `username` | `GITHUB_ACTOR`          | Username for the token                                  |
| `debug`    | `false`                 | Enable debug logging                                    |

## Outputs

| Output             | Description                                       |
| ------------------ | ------------------------------------------------- |
| `changes-detected` | `true` if a commit was made                       |
| `commit`           | SHA of the new commit (or HEAD if nothing changed) |

## Example

```yaml
- uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@master
- run: ./generate.sh
- uses: https://git.mouse-lake.ts.net/actions/toolkit/git-commit@master
  with:
    message: Regenerate files
    files: generated/
```
