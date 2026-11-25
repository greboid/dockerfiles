# toolkit

Actions for [Forgejo Runner](https://forgejo.org/docs/latest/user/actions/). Each action runs as a container, so no Node.js or other setup is needed on the runner. See each action's README for its inputs, outputs and an example.

## Actions

| Action                                  | Description                                                   |
| --------------------------------------- | ------------------------------------------------------------- |
| [`git-checkout`](git-checkout/)         | Check out a repository                                        |
| [`git-commit`](git-commit/)             | Commit workspace changes and optionally push                  |
| [`git-try-push`](git-try-push/)         | Push with retries, rebasing between attempts                  |
| [`release`](release/)                   | Create a Forgejo release and upload assets                    |
| [`go-setup`](go-setup/)                 | Install a Go toolchain                                        |
| [`go-build`](go-build/)                 | Build a static Go binary                                      |
| [`go-test`](go-test/)                   | Run `go test`                                                 |
| [`go-vet`](go-vet/)                     | Run `go vet`                                                  |
| [`go-lint`](go-lint/)                   | Check Go files are `gofmt` formatted                          |
| [`node-setup`](node-setup/)             | Install Node.js and pnpm/yarn/npm                             |
| [`docker-login`](docker-login/)         | Log in to a container registry                                |
| [`docker-build`](docker-build/)         | Build a container image                                       |
| [`docker-push`](docker-push/)           | Build and push a container image                              |
| [`package-latest`](package-latest/)     | Find the latest and next version of a registry package        |
| [`apko-changelog`](apko-changelog/)     | Changelog message from two apko lock files                    |

## Usage

Reference an action by its directory:

```yaml
steps:
  - uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@master
  - uses: https://git.mouse-lake.ts.net/actions/toolkit/go-test@master
```
