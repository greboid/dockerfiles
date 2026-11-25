# docker-build

Builds a container image without pushing it. A tagged single-platform build is loaded into the runner's docker engine so later steps can use it.

## Inputs

| Input        | Default | Description                                                                 |
| ------------ | ------- | --------------------------------------------------------------------------- |
| `tags`       |         | Image references to tag the image with (comma or whitespace separated)      |
| `context`    | `.`     | Build context directory                                                     |
| `file`       |         | Containerfile/Dockerfile path, relative to the context                      |
| `target`     |         | Build target stage                                                          |
| `platforms`  |         | Target platforms, e.g. `linux/amd64,linux/arm64`. With more than one, nothing is loaded into the engine |
| `build-args` |         | Build arguments, one `KEY=VALUE` per line                                   |
| `no-cache`   | `false` | Do not use the build cache                                                  |
| `debug`      | `false` | Enable debug logging                                                        |

## Outputs

| Output    | Description           |
| --------- | --------------------- |
| `imageid` | ID of the built image |

## Example

```yaml
- uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@master
- uses: https://git.mouse-lake.ts.net/actions/toolkit/docker-build@master
  with:
    tags: myapp:local
- run: docker run --rm myapp:local --version
```
