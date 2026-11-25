# docker-login

Replacement for `docker/login-action` on [Forgejo Runner](https://forgejo.org/docs/latest/user/actions/):
a docker action — a small Go binary in its own container, built from this repository by the runner —
that logs credentials into the docker config later steps find and exports `DOCKER_CONFIG`. No Node.js,
no `actions/node20` runtime. `buildah login` (baked into the action image) verifies the credentials
against the registry (Registry HTTP API v2) and writes the config.

## Usage

```yaml
jobs:
  build:
    runs-on: alpine # a label backed by a docker:// image (alpine works)
    steps:
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@v1
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/docker-login@v1
        with:
          registry: forgejo.example.com
          username: ${{ secrets.REGISTRY_USER }}
          password: ${{ secrets.REGISTRY_PASSWORD }}
```

(The `https://` prefix form is unambiguous; bare `owner/repo@ref` also works when the instance
`DEFAULT_ACTIONS_URL` is set to it.)

The action container needs network access to the registry.

### Login via a Forgejo Authorized Integration

On Forgejo v16+, instead of a shared secret you can log in with a JWT minted for an
[Authorized Integration](https://forgejo.org/docs/latest/user/api/authorized-integrations/)
(the workflow itself is the credential — Forgejo checks the JWT's repository, workflow, ref and
event claims against the integration's rules):

```yaml
enable-openid-connect: true # workflow top level; makes the runner expose a token endpoint

jobs:
  build:
    steps:
      - uses: https://git.mouse-lake.ts.net/actions/toolkit/docker-login@v1
        with:
          audience: my-registry-integration # logs in to this Forgejo instance's registry
```

### Zero-config mode

With no `registry`, `username` or `password` input at all, the action logs in to the Forgejo
instance the workflow runs on, using the job's `GITHUB_TOKEN` as the password and the repository
owner as the username. That is enough to push to the instance's own registry (see `docker-push`).

## Inputs

| input      | description                                                                                                                           |
| ---------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| `registry` | Registry to log in to; with no inputs at all, defaults to the Forgejo instance of this workflow using `GITHUB_TOKEN`                   |
| `username` | Registry username (ignored in authorized-integration mode; Forgejo reads the token from the password)                                 |
| `password` | Registry password or token; takes precedence over `audience`                                                                          |
| `audience` | Audience of a Forgejo Authorized Integration; when set (and `password` is not), fetches an OIDC JWT from the runner and logs in to the Forgejo instance of this workflow with it (requires Forgejo v16+ and `enable-openid-connect: true`) |
| `ghcr`     | GitHub PAT with `write:packages` to also log in to ghcr.io (for `docker-push`'s ghcr mirror)                                          |
| `debug`    | Enable debug logging                                                                                                                  |

## Outputs

None.

## Notes

- The credentials land in `<workspace>/.actions/docker/config.json` (or the `DOCKER_CONFIG`
  directory a previous step chose); `DOCKER_CONFIG` is exported for subsequent steps. If you don't
  want credentials persisted after the job, clean up in a final step.
- For Docker Hub the config key is rewritten to the legacy `https://index.docker.io/v1/` location,
  so the docker CLI finds the credentials too (buildah reads both spellings).
