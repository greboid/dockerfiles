# git-commit

Commit the changes `git-checkout` left in the workspace — all of them, or only files matching
the given pathspecs — and optionally push the commit back to its branch. A compact replacement
for `stefanzweifel/git-auto-commit-action` on
[Forgejo Runner](https://forgejo.org/docs/latest/user/actions/): a docker action — a small Go
binary in a container with git baked in, built from this repository by the runner — driving the
git CLI. When nothing ends up staged the step succeeds with `changes-detected=false` and pushes
nothing.

## Usage

```yaml
jobs:
  update:
    runs-on: alpine # a label backed by a docker:// image (alpine works)
    steps:
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@v1
      - run: ./generate.sh
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/git-commit@v1
        with:
          message: regenerate
          files: generated/*.txt
```

(The `https://` prefix form is unambiguous; bare `owner/repo@ref` also works when the instance
`DEFAULT_ACTIONS_URL` is set to it.)

### Authentication

The push authenticates like the other git-* actions: the `token` input, an authorized
integration ID token when `audience` is given (requires `enable-openid-connect: true`), then
`GITHUB_TOKEN`, then `GITEA_TOKEN`. The token lives only in the environment for the push —
never in `.git/config`, argv or workflow logs — and is masked.

## Inputs

| input      | description                                                                                    |
| ---------- | ---------------------------------------------------------------------------------------------- |
| `message`  | Commit message (required)                                                                       |
| `files`    | Whitespace-separated pathspecs to stage, with git's own globbing; default stages every change, deletions included |
| `name`     | Committer name; defaults to `GITHUB_ACTOR`                                                      |
| `email`    | Committer email; defaults to `<name>@users.noreply.<instance host>`                             |
| `push`     | Push the commit to origin after committing (default `true`)                                     |
| `branch`   | Branch to push to; defaults to the checked-out branch                                           |
| `token`    | Access token for pushing; defaults to the audience input's ID token, then `GITHUB_TOKEN`, then `GITEA_TOKEN` |
| `audience` | Audience of a Forgejo Authorized Integration; used as the auth token when token is not set      |
| `username` | Username for token basic auth; defaults to `GITHUB_ACTOR` then `x-access-token`                 |
| `debug`    | Enable debug logging                                                                            |

## Outputs

| output            | description                                                            |
| ----------------- | ---------------------------------------------------------------------- |
| `changes-detected`| `true` when a commit was created, `false` when there was nothing to commit |
| `commit`          | The SHA of the created commit (or of HEAD when nothing changed)         |

## Notes

- Run `git-checkout` first: the workspace must contain a `.git` directory.
- The pathspecs are delivered verbatim to `git add`, so git's own globbing applies.
