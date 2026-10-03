# build-common

Fleet-wide documentation for the `containers` image organisation: one repo
per image, depbot update PRs, `YYYY.M.D-alpineX.Y` tags plus `:latest`.

This repo no longer hosts shared workflows. Each image repo carries its own
`.forgejo/workflows/build.yml` (build + publish on push to `master`) and
`.forgejo/workflows/pr.yml` (build-only on pull requests); the pipeline is
small enough that per-repo beats indirection. The repos are otherwise
identical in shape:

- **`apko.yaml`** — the image: exact-pinned packages from the `packages`
  organisation's Alpine repository (the same repository the packages org
  publishes to with melange). Nothing is compiled in an image build; the
  repo only shapes the image around packages that carry the binaries
  (entrypoint, run-as user, environment, volumes, paths).
- **`keys/packages@<hash>.rsa.pub`** — the org repository's index key. apk
  and apko match an index signature to a key by file name only — the
  signature embeds `<owner>@<sha256(DER)>.rsa.pub` — so the key is
  committed under its hash-derived name (the endpoint's own basename,
  `key`, can never match). If the Forgejo registry key ever rotates:
  ```sh
  curl -fsSL -o key.rsa.pub https://git.mouse-lake.ts.net/api/packages/packages/alpine/key
  hash=$(openssl pkey -pubin -in key.rsa.pub -outform DER | sha256sum | cut -d' ' -f1)
  mv key.rsa.pub "keys/packages@${hash}.rsa.pub"
  git rm keys/packages@<old-hash>.rsa.pub   # and update the path in apko.yaml
  ```
- **`build.yml`** — checkout → install the committed key for the host's apk
  and `apk add apko` from the org repository (the build tool is just another
  org package, the same patched apko that published the packages the image
  pins) → docker-login (job token; zero inputs) → `apko publish apko.yaml`
  pushing `:YYYY.M.D-alpine3.24` and `:latest`. Tags are overwritten on
  every build; there is no immutability guard. The three mirrored images
  (`alpine`, `dsp`, `golang`) additionally push the same build to
  `ghcr.io/greboid/dockerfiles` when `GH_USERNAME`/`GH_TOKEN` are
  configured — one `apko publish` pushes to every listed tag, and the
  mirror skips quietly without credentials.
- **`pr.yml`** — same trust + apko install, then `apko build` to a local
  tarball: no login, no push.
- **`.depbot.yaml`** — identical in every repo:
  ```yaml
  min_age_days: 0

  automerge:
    - ecosystems: [apk]
  ```
  `min_age_days: 0` everywhere (decision 7). depbot's apk ecosystem bumps the exact package
  pins in `apko.yaml` when the packages org publishes newer versions; the
  app version itself is tracked in the packages org.

## Local builds

The repository key is committed under `keys/`, so a plain apko build works
from the repo root (apko comes from the org repository — `apk add apko`
after trusting it as in `build.yml` — or from the containertools releases):

```sh
apko build apko.yaml myimage:local image.tar
```

## Retirement note

`base`, `baseroot` and `git` built the shared rootfs chain that apko images
no longer use (accounts and baselayout come from the packages instead). They
were not converted and should be retired.
