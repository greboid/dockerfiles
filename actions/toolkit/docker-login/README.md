# docker-login

Logs in to a container registry so later steps (such as `docker-push`) can use the credentials. With no inputs it logs in to this Forgejo instance's registry using the job token.

## Inputs

| Input      | Default | Description                                                                  |
| ---------- | ------- | ---------------------------------------------------------------------------- |
| `registry` |         | Registry to log in to; defaults to this Forgejo instance                     |
| `username` |         | Registry username                                                            |
| `password` |         | Registry password or token                                                   |
| `audience` |         | Forgejo Authorized Integration audience to log in with instead of a password (needs `enable-openid-connect: true`) |
| `ghcr`     |         | GitHub token with `write:packages` to also log in to ghcr.io                 |
| `debug`    | `false` | Enable debug logging                                                         |

## Outputs

None.

## Example

```yaml
- uses: https://git.mouse-lake.ts.net/actions/toolkit/docker-login@master
  with:
    registry: registry.example.com
    username: ${{ secrets.REGISTRY_USER }}
    password: ${{ secrets.REGISTRY_PASSWORD }}
```
