# git-checkout

Pure-Go replacement for `actions/checkout` on [Forgejo Runner](https://forgejo.org/docs/latest/user/actions/):
a small static Go binary in an Alpine image that fetches the workflow repository over git smart
HTTP using go-git — no Node.js, no `actions/node20` runtime, no git binary. It talks to the runner
through the standard `GITHUB_*` files (`-flag` args from the action inputs, `$GITHUB_OUTPUT`).

## Usage

```yaml
jobs:
  build:
    runs-on: docker # a label backed by a docker:// image
    steps:
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@v1
        with:
          token: ${{ secrets.GITHUB_TOKEN }}
```

(The `https://` prefix form is unambiguous; bare `owner/repo@ref` also works when the instance
`DEFAULT_ACTIONS_URL` is set to it.)

The image is built from this repository by the images workflow and pushed to the instance
registry. The `vendor/` tree is committed at the repository root, so that build works without a
Go module proxy.

## Behaviour

Always performs a fresh clone into `path` (contents of the destination are wiped first, including
the workspace itself). Refs are tried in order: the given `ref` (default `GITHUB_REF`), then for
pull requests the `/head` and `/merge` variants plus `GITHUB_HEAD_REF` (Forgejo uses
`refs/pull/<n>/head`, GitHub uses `/merge`; both work), then bare-name and `GITHUB_SHA`
interpretations. `fetch-depth: 0` fetches full history and tags. Branch refs are checked out as
local branches; tags and PRs as a detached HEAD. HTTPS + token only (`x-access-token`-style basic
auth, exactly what `actions/checkout` sends on GitHub, which Forgejo accepts); SSH is not
supported. Outputs: `ref`, `commit`.

## Limitations

SSH keys, filters and sparse checkouts are not supported, and it always re-clones rather than
cleaning in place.

## Development

```sh
make check   # go test + go vet (from the repository root)
make images  # build the action image locally, as the images workflow does (from the repository root)
```
