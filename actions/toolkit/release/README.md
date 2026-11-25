# release

Creates a release on this Forgejo instance and uploads files to it. If a release for the tag already exists, it is replaced.

## Inputs

| Input        | Default            | Description                                                         |
| ------------ | ------------------ | ------------------------------------------------------------------- |
| `tag`        | the pushed tag     | Tag to release; created if it doesn't exist                         |
| `name`       | the tag            | Release title                                                       |
| `body`       |                    | Release notes                                                       |
| `commitish`  | current commit     | Commit to create the tag at                                         |
| `assets`     |                    | Files to upload, one per line; `name=path` to rename                |
| `replace`    | `true`             | Replace an existing release for the tag                             |
| `draft`      | `false`            | Mark as draft                                                       |
| `prerelease` | `false`            | Mark as pre-release                                                 |
| `repository` | this repository    | Repository to release into (`owner/name`)                           |
| `token`      | `GITHUB_TOKEN`     | Access token                                                        |
| `audience`   |                    | Forgejo Authorized Integration audience to use instead of a token (needs `enable-openid-connect: true`) |
| `debug`      | `false`            | Enable debug logging                                                |

## Outputs

| Output   | Description                 |
| -------- | --------------------------- |
| `id`     | Release ID                  |
| `url`    | Release URL                 |
| `assets` | Number of assets uploaded   |

## Example

```yaml
- uses: https://git.mouse-lake.ts.net/actions/toolkit/release@master
  with:
    tag: v1.2.3
    body: Release notes
    token: ${{ secrets.RELEASE_TOKEN }}
    assets: |
      bin/tool
      bin/tool.sha256
```
