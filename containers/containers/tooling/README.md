# Tooling: our own melange, apko and oras builds

A single script builds known-good upstream releases with our patch sets
applied. Nothing is downloaded from a registry first: CI and local builds
both run `tooling/tools.sh` and get the same binaries.

## Usage

```sh
tooling/tools.sh          # all three -> tooling/bin/
tooling/tools.sh oras     # just one
```

Requirements: git, go, network. Output binaries land in `tooling/bin/`
(gitignored).

## Versions and patches

- `versions.env` — the known-good upstream tags. Bump deliberately: a new
  tag may need patch rebases.
- `melange/patches/`, `apko/patches/` — our patch sets, applied in lexical
  order; see each directory's README for what they do and rebase notes.
  oras is used as upstream ships it (no patches).

## Pipeline use

The ci jobs apk add `bubblewrap` next to uv: it is melange's default
container runner for builds, and the alpine job image does not ship it.
The jobs call `tooling/tools.sh` to build the binaries, then
use `tooling/bin/<tool>`. There is no published-binary step: every build is
fresh from source at the recorded version.

The uv scripts (this directory's `*.py`) carry a `#!/usr/bin/env -S uv run
--script` shebang for direct local use. Alpine job images ship BusyBox, whose
`env` has no `-S`, so CI invokes them as `uv run --script tooling/<script>`;
test-release-workflows.py enforces that for every run block in every
workflow file.

`tooling/single-package.py <name>` wraps melange for local package builds
with the same flags as the ci packages jobs (epoch, namespace,
license, repo URL), writing to `dist/x86_64/`:

```sh
tooling/single-package.py <name>                  # signed local build -> dist/
tooling/single-package.py <name> --key keys/release.rsa --out-dir dist/release
                                                  # CI-style: per-run key, isolated repo
tooling/single-package.py <name> --remote-repo URL --remote-key forgejo.rsa.pub
                                                  # fan-out child: only <name> is built,
                                                  # deps resolve from the published repo
```

It plans build dependencies from the recipes with the same graph
`all-packages.py` uses, building any missing ones into the same signed
repository first. In remote mode (`--remote-repo` with `--remote-key`) it
builds only the target recipe: affected dependencies were already awaited
in the published repository and unaffected ones are published from earlier
runs, so each recipe is built exactly once per fan-out. The ci pull-request
jobs run the local builder with a throwaway key and an isolated
`--out-dir dist/release`; the push-path package legs of the ci matrix run the
remote builder with the organisational key.

`tooling/await-published.py` is the fan-out's dependency barrier: given a
target recipe (or a `--list` of recipes, as image legs use for their
closure), it computes the transitive local package set, intersects
it with the affected selection, and polls the published Forgejo Alpine
repository for each expected `name-version-r epoch.apk` until the owning
legs publish or a timeout marks the leg failed. The waves are computed
on the fly — the recipe graph is the input, never workflow YAML.

`tooling/check-bumps.py` rejects recipe changes that do not bump version
or epoch before a push fan-out is dispatched: published file names are the
freshness signal the awaits rely on, and an unbumped recipe would
republish the same name. The comparison is semantic (comments and
formatting changes are exempt); a missing or unusable base skips the
check. republish has no base and never runs it.

`tooling/all-packages.py` discovers every `packages/*.yaml` recipe and
builds them sequentially in local dependency order using the same builder.
It prioritizes `os-release`, includes the runtime dependencies of local build
inputs (including named subpackages), rejects build dependency cycles, and stops
on failure. External dependencies remain melange's responsibility. Requires
[uv](https://docs.astral.sh/uv/getting-started/installation/), which automatically
manages Python and inline-pinned PyYAML; no manual virtualenv or system Python
packages are needed. Use `--dry-run` to print the order without building.
This builds packages only, not images.

`tooling/all-images.py` discovers `images/*.yaml` and builds each image in
alphabetical order via `tooling/single-image.sh`, stopping on the first failure.
It uses existing packages in `dist/`; it does not rebuild packages or publish
images. Output lands in `dist/images/`. Both Python scripts have a uv shebang
and can be executed directly:

```sh
./tooling/all-packages.py           # build packages first
./tooling/all-images.py --dry-run   # list images without building
./tooling/all-images.py            # build all images
```
