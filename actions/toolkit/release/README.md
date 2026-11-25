# release

Create a release on the Forgejo instance the workflow runs on and upload
files as its assets — a compact replacement for `softprops/action-gh-release`
on [Forgejo Runner](https://forgejo.org/docs/latest/user/actions/): a docker
action — a small Go binary in its own container, built from this repository by
the runner. No Node.js, no `actions/node20` runtime.

The release is created in the workflow's repository (or the `repository`
input's). When the tag does not exist yet it is created at the commit the
workflow runs on, or at `commitish`. When a release for the tag already
exists — a rerun of the same build, most commonly — it is replaced wholesale:
the old release and its assets are deleted and the new ones uploaded, so
stale assets can never survive. The tag itself is kept across replacements.

## Usage

```yaml
enable-openid-connect: true # workflow top level; makes the runner expose a token endpoint

jobs:
  publish:
    runs-on: alpine
    steps:
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@v1
      - run: ./build.sh
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/release@v1
        with:
          audience: my-release-integration # Forgejo Authorized Integration audience
          tag: v1.2.3
          body: "Release notes"
          assets: |
            bin/tool
            bin/tool.sha256
```

(The `https://` prefix form is unambiguous; bare `owner/repo@ref` also works when the instance
`DEFAULT_ACTIONS_URL` is set to it.)

Reruns of the same tag replace the release, which makes the step
idempotent for retried builds.

### Authentication via a Forgejo Authorized Integration

On Forgejo v16+, prefer a
[Authorized Integration](https://forgejo.org/docs/latest/user/api/authorized-integrations/)
over a stored token: the workflow itself is the credential — Forgejo checks the JWT's
repository, workflow, ref and event claims against the integration's rules, and the
integration is configured with the capabilities it grants.

```yaml
enable-openid-connect: true # workflow top level

jobs:
  publish:
    steps:
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/release@v1
        with:
          audience: my-release-integration # safe to commit -- not confidential
          tag: v1.2.3
```

The credential resolution order is the `token` input, then the `audience`
input's ID token, then the job's `GITHUB_TOKEN`. The job token usually
lacks the capabilities to create releases; it is the fallback of last
resort.

## Inputs

| input       | description                                                                                                                                    |
| ----------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `tag`       | Tag to release; created at `commitish` when it does not exist yet. Defaults to the pushed tag's name when the workflow runs on a tag event       |
| `name`      | Release title; defaults to the tag                                                                                                              |
| `body`      | Release notes                                                                                                                                   |
| `commitish` | Commit to point the tag at; defaults to the commit the workflow runs on                                                                          |
| `assets`    | Whitespace/newline-separated files to upload (workspace-relative), each as `path` or `name=path` to rename; duplicate names are an error          |
| `replace`   | Replace an existing release for the tag, its assets included (the tag itself is kept); default `true`                                            |
| `draft`     | Mark the release as a draft                                                                                                                     |
| `prerelease`| Mark the release as a pre-release                                                                                                               |
| `repository`| owner/name to release into; defaults to the workflow's repository                                                                                |
| `token`     | Access token for the Forgejo API; takes precedence over `audience`, then the job token                                                           |
| `audience`  | Audience of a Forgejo Authorized Integration; when set (and `token` is not), fetches an OIDC JWT from the runner and authenticates with it (requires Forgejo v16+ and `enable-openid-connect` set to true in the workflow) |
| `debug`     | Enable debug logging                                                                                                                            |

## Outputs

| output   | description                     |
| -------- | ------------------------------- |
| `id`     | The release's numeric id        |
| `url`    | The release's HTML URL          |
| `assets` | Number of assets uploaded       |

## Notes

- Assets upload under their basename unless renamed with `name=path`;
  names must not contain path separators and must be unique within the
  release.
- Assets download from the stable
  `/releases/download/<tag>/<asset>` URL once the release is published.
- Replacing keeps the tag where it is; re-pointing an existing tag is a
  `git push -f` concern, not this action's.
- The action container needs network access to the Forgejo instance.
