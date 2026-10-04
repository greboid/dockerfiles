Container image `git.mouse-lake.ts.net/containers/pocket-id` — built with apko
from the `packages` organisation's Alpine repository: the image installs the
org's `pocket-id` package at an exact pinned version, so the binary,
license notices and version bumps live in `packages/pocket-id`
(upstream: https://github.com/stonith404/pocket-id).

- Tags: `:YYYY.M.D-alpineX.Y` and `:latest`, pushed on every `master` build
- Builds run on push to `master` via this repo's `.forgejo/workflows`
- Package version pins in `apko.yaml` arrive as depbot PRs (`.depbot.yaml`);
  the app version itself is tracked by depbot in the `packages` org's
  pocket-id repo

## Runtime notes

- Listens on `0.0.0.0:1411`; state lives in `/data` (mount a volume there).
  The working directory is `/`, so the app's relative `data/` paths land in
  `/data` by default.
- `ENCRYPTION_KEY` must be set to at least 16 bytes (required by upstream).
- `/healthz` answers 204 for load-balancer or compose healthchecks; the
  binary also ships a `pocket-id healthcheck` command.
- Configuration is upstream's environment-variable set (`APP_URL`, `PORT`,
  SMTP_*, LDAP_*, ...); nothing is baked into the image beyond defaults.

## Local build

The repository key is committed under `keys/`, so a plain apko build works:

```sh
apko build apko.yaml pocket-id:local image.tar
```
