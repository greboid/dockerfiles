# packages/common: shared CI for the packages organisation

One pipeline for every package repo. A repo's own workflow is a thin stub
that calls the reusable workflow here:

```yaml
name: build

on:
  push:
    branches: [main]
  pull_request:
  workflow_dispatch:

jobs:
  ci:
    uses: packages/common/.forgejo/workflows/build-package.yml@main
    secrets: inherit
```

The caller job stays bare: `uses:` and `secrets: inherit`, nothing else.
A `runs-on` on a `uses:` job is not valid reusable-workflow syntax, and
Forgejo blocks such runs server-side with "workflow was not executed due
to an error that blocked the execution attempt", dispatching no jobs at
all - which is exactly what the long line of failed redis runs was. The
`runs-on: alpine` belongs on the job inside the called workflow, where
this instance's reusable-workflow expansion works fine (the containers
org builds the same way against `containers/build-common`).

`secrets: inherit` forwards the caller's secrets to the called workflow.
`APK_SIGNING_KEY` and `RELEASE_PACKAGE_TOKEN` are organisation secrets,
so every repo in the org has them; no per-repo secret configuration
exists or is needed.

Inside a called workflow `github.repository` and `github.event` are the
caller's, so package name, origin URL and publish gating derive per repo:
pushes to the caller's main branch build, test, publish and verify; pull
requests build and test only.

## How builds run

The organisation's Alpine repository (`api/packages/packages/alpine`) is
the single home for org packages: builds resolve dependencies from it and
publish back into it. The workflow

1. installs `bubblewrap` and `openssl` from Alpine;
2. trusts the repository's index key (fetched from the org-level key
   endpoint) in `/etc/apk/keys`, and appends the repository to
   `/etc/apk/repositories`;
3. installs `melange` from that repository with `apk` - the build tool is
   just another org package, the same one that seeded the repository;
4. writes the org signing key (refusing to publish without it; PR runs
   without the secret get a throwaway key and publish nothing);
5. builds the repo's `melange.yaml`, resolving org-package dependencies
   from the same repository (melange is given the repository and its
   index key explicitly, since it matches index signatures by key file
   name);
6. runs the recipe's tests against the freshly built package;
7. on main pushes only, uploads the signed apk to the repository and
   verifies the round trip against the published copy.

Rebuilding a bootstrap package (os-release, go, melange, apko) after a
version bump needs no special handling: its repo builds like any other,
with the previously published melange doing the building.

## Bootstrapping the org repository

`bootstrap.yml` is a standalone workflow, run by hand from the Actions
tab (Run workflow) only when the org's Alpine repository has no usable
tools - in practice once, before the first package build. The builder is
upstream melange installed with `go install` (not packaged in Alpine
3.24). The workflow clones the org's own package repos - `os-release`,
`go`, `melange`, `apko` - and walks them as a DAG, building each repo's
real `melange.yaml` in dependency order and publishing the signed
packages. The bootstrap has been run; from then on every repo in the
organisation - those four included - builds through the reusable
workflow above.

## Layout

- `.forgejo/workflows/build-package.yml` — the reusable build pipeline
- `.forgejo/workflows/bootstrap.yml` — hand-run org repository seeding
