# git-try-push

Push the checked-out branch (and optional tags) to the remote with retries — when a concurrent
push won the race, the remote branch is fetched and the local commits rebased onto it before
waiting and retrying. A compact replacement for `cutlerydrawer/action-git-try-push` on
[Forgejo Runner](https://forgejo.org/docs/latest/user/actions/): a docker action — a small Go
binary in a container with git baked in, built from this repository by the runner — driving the
git CLI, built for matrix jobs and pipelines that all push to the same branch.

## Usage

```yaml
jobs:
  release:
    runs-on: alpine # a label backed by a docker:// image (alpine works)
    steps:
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@v1
      - run: ./bump-version.sh
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/git-commit@v1
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/git-try-push@v1
        with:
          tags: v1.2.3
          push_flags: --force-with-lease
```

(The `https://` prefix form is unambiguous; bare `owner/repo@ref` also works when the instance
`DEFAULT_ACTIONS_URL` is set to it.)

### Retry behaviour

Each failed attempt fetches the remote branch and rebases the local commits onto it
(`--autostash`), then waits — 2s, growing by the factor each time, capped at 60s — and pushes
again, up to `retries` times. A conflicted rebase fails the step (a human has to resolve it);
a failed fetch only logs a warning and retries the push as-is. The fetch updates the
remote-tracking ref, so a user-given `--force-with-lease` has a current expectation.

### Authentication

Same chain as `git-commit`: the `token` input, an authorized integration ID token when
`audience` is given (requires `enable-openid-connect: true`), then `GITHUB_TOKEN`, then
`GITEA_TOKEN`. The token lives only in the environment — never in `.git/config`, argv or
workflow logs — and is masked.

## Inputs

| input          | description                                                                       |
| -------------- | --------------------------------------------------------------------------------- |
| `branch`       | Branch to push HEAD to; defaults to the checked-out branch (none when detached)   |
| `tags`         | Whitespace-separated tag names to push alongside the branch                       |
| `remote`       | Remote to push to (default `origin`)                                              |
| `retries`      | How many times to retry after a failed push (default `10`)                        |
| `initial_wait` | Seconds to wait before the first retry (default `2`)                              |
| `wait_factor`  | Multiplier applied to the wait after each retry (default `2`)                     |
| `wait_max`     | Cap, in seconds, of the growing wait (default `60`)                               |
| `rebase`       | Catch up with the remote branch (fetch + rebase --autostash) between attempts (default `true`) |
| `push_flags`   | Extra whitespace-separated git push flags, e.g. `--force-with-lease`              |
| `token`        | Access token; defaults to the audience input's ID token, then `GITHUB_TOKEN`, then `GITEA_TOKEN` |
| `audience`     | Audience of a Forgejo Authorized Integration; used as the auth token when token is not set |
| `username`     | Username for token basic auth; defaults to `GITHUB_ACTOR` then `x-access-token`   |
| `debug`        | Enable debug logging                                                              |

## Outputs

| output | description                                      |
| ------ | ------------------------------------------------ |
| `head` | The SHA that was pushed (HEAD after any rebase)  |

## Notes

- Run `git-checkout` first: the workspace must contain a `.git` directory.
- Pushing a detached HEAD with no `branch` and no `tags` input fails; give one or the other.
- A fallback author/committer identity (the workflow actor at the instance's noreply domain) is
  exported for the rebase, so replaying commits in the fresh container works without git config.
