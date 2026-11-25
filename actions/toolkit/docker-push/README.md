# docker-push

Builds a container image and pushes it to a registry. Without `tags`, tags are derived from the workflow ref:

| Ref                        | Tags                                              |
| -------------------------- | ------------------------------------------------- |
| `master` / `main` branch   | `dev`                                             |
| `v1.2.3` tag               | `1.2.3`, `1.2`, `1`, and `latest` for the newest major |

Run `docker-login` first, or pass credentials directly.

## Inputs

| Input             | Default            | Description                                                          |
| ----------------- | ------------------ | -------------------------------------------------------------------- |
| `tags`            |                    | Image references to push (comma or whitespace separated)             |
| `registries`      | this instance      | Registries for derived tags (comma separated)                        |
| `repository`      | `owner/repo`       | Image name for derived tags                                          |
| `token`           | `GITHUB_TOKEN`     | Token used to list repository tags for the `latest` rule            |
| `context`         | `.`                | Build context directory                                              |
| `file`            |                    | Containerfile/Dockerfile path, relative to the context               |
| `target`          |                    | Build target stage                                                   |
| `platforms`       |                    | Target platforms, e.g. `linux/amd64,linux/arm64`                     |
| `build-args`      |                    | Build arguments, one `KEY=VALUE` per line                            |
| `push`            | `true`             | Push after building                                                  |
| `no-cache`        | `false`            | Do not use the build cache                                           |
| `registry`        |                    | Registry for the credentials below                                   |
| `username`        |                    | Registry username                                                    |
| `password`        |                    | Registry password                                                    |
| `ghcr`            | `false`            | Also push to `ghcr.io/owner/repo`                                    |
| `ghcr-repository` | `owner/repo`       | GHCR repository to push to                                           |
| `ghcr-token`      |                    | GitHub token with `write:packages`                                   |
| `ghcr-username`   | `GITHUB_ACTOR`     | GHCR username                                                        |
| `debug`           | `false`            | Enable debug logging                                                 |

## Outputs

| Output    | Description                                                  |
| --------- | ------------------------------------------------------------ |
| `imageid` | ID of the built image                                        |
| `digest`  | Digest of the pushed image                                   |
| `version` | The version for tag refs, the commit SHA on master/main      |
| `tags`    | The image references pushed, comma separated                 |

## Example

```yaml
- uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@master
- uses: https://git.mouse-lake.ts.net/actions/toolkit/docker-login@master
- uses: https://git.mouse-lake.ts.net/actions/toolkit/docker-push@master
  with:
    platforms: linux/amd64,linux/arm64
```
