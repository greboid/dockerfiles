# Containers

Source-built APK packages and minimal apko images, published to Forgejo with
our own builds of standard tooling: **melange** builds packages, **apko**
builds images, **oras** handles registry copying and mirroring. Release
builds (pushes to main, `republish.yml`) sign packages with the
organisational `APK_SIGNING_KEY`; pull requests and local builds use
gitignored throwaway keys whose build/test guests verify signatures. The
Forgejo repository index is signed by Forgejo itself: publication integrity
comes from the HTTPS round trip, exact package pins, and content-addressed
image references.

> Historical note: an earlier, much larger gate-based design (git-ref
> release locks, detached provenance, self-built APK/index verifiers) was
> replaced by this setup; that history is in the git log.

## Layout

| Path | Purpose |
| --- | --- |
| `packages/` | melange package recipes (signed for build/test guests) |
| `images/` | apko image definitions (consume published packages) |
| `tooling/` | builds our own melange/apko/oras from source with our patches ([README](tooling/README.md)) |
| `.forgejo/workflows/` | the whole pipeline |

Application recipes pair `packages/<name>.yaml` (melange) with
`images/<name>.yaml` (apko). Application binaries and bundled assets belong
in the APK, not an image build stage. Runtime dependencies currently come
from Alpine, following the Miniflux pattern.

| Image | Runtime notes |
| --- | --- |
| `miniflux` | External Postgres via `DATABASE_URL` |
| `golink` | Persist `/home/nonroot`; default database `/home/nonroot/golink.db` |
| `ergo` | Mount configuration/data at `/ircd`; default config `/ircd/ircd.yaml`; language assets at `/ircd-bin/languages` |
| `httpredirect` | HTTP on port 8080; configure redirects with `TARGETHOST`, `TARGETPATH`, `EXCLUDEQUERY`, `HTTPREDIRECT`, `TEMPREDIRECT`, and `PORT` |
| `identd` | Ident on port 8000; configure with `-port`/`-response` or `PORT`/`RESPONSE` |
| `dockercleanup` | Prunes stopped containers and unused images via the Docker API; `DURATION` below one minute runs once, `ALLIMAGES=true` also prunes non-dangling images, `CONTAINERS=false` skips containers |
| `webhooked` | Ensures a GitHub webhook (needs `TOKEN` and `URL`) across a user's repositories; `MONITOR` below one minute runs once; exits 0 with usage when unconfigured |
| `registryauth` | Token auth server with embedded registry on port 8080 (`PORT`); configure with `USERS` (YAML map of bcrypt hashes), `PUBLIC`, `REALM`, `ISSUER`, `SERVICE`; certs/storage under `/data` (writable mount required) |
| `goplum` | Health checker loading `goplum.conf` (`-config`); mTLS gRPC API on 7586 (`API_PORT`); tombstone at `/tmp/goplum.tomb` (`TOMBSTONE`) |
| `puzzles` | Word/image puzzle helpers on :8080 (fixed); wordlists bundled at `/app/wordlists` (`WORDLIST_DIR`) |

The table above covers the first four batches; all 39 image definitions are
in `images/`: 36 application images running as `65532:65532` (writable
mounts must allow that UID/GID), plus three reusable build/CI images that
keep the legacy rootful, shell-CMD interface: `golang` (Go 1.27.1,
go-licenses 2.0.1), `rust` (Rust 1.98.1 via rustup 1.29.1), and
`alpine-runner` (Buildah 1.45.1, Forgejo Actions job environment). Node is
out of scope for now, at the user's request. See [migration
status](MIGRATION.md) for later batches and the remaining cutover work.

## Pipeline

`.forgejo/workflows/ci.yml` runs on every pull request and every push to
main. Every path re-derives the affected set with `tooling/affected.py`
(changed files mapped onto the recipe dependency graph): a changed recipe
selects itself and every recipe that transitively build-depends on it; an
image is selected when its definition changed or any local package it
installs is affected; tooling changes select everything (melange-side
tooling all packages, apko-side all images, oras nothing). A base CI cannot
anchor to — an empty or all-zeros comparison commit, an unknown ref,
disjoint histories — selects everything: selective builds are only ever an
optimization, never a correctness input. The selection validates the recipe
graph (cycles, duplicate providers, name mismatches) and every image
definition before anything builds.

The workflow fans out natively: the **plan** job runs the offline checks,
the version bump check and the selector. On **publish runs** (push to
main, manual dispatch) it then generates the publish DAG with
`tooling/gen_dag.py` — one job per affected package, one per affected
image — commits it to a per-run branch (`ci/dag-<run_id>`) and pushes it;
the push triggers the generated workflow. Its `needs` edges ARE the
dependency barrier: a job is dispatched only once everything it consumes
has been published by the jobs it needs, so a leg never holds a runner
slot while waiting (the previous ready-poll held one and deadlocked a
capacity-2 runner), and a failed leg skips its dependents instantly
instead of failing them after a 30-minute polling timeout. The edges
cannot live in a committed workflow — needs is resolved at scheduling
time over statically-named jobs, while the node set is the per-commit
affected list — which is why the workflow is generated; a cleanup job
deletes the branch afterwards. On **pull requests** the plan job emits
the fan-out as two JSON matrices instead; package legs build the affected
closure locally (throwaway key, nothing published, no secrets) and image
legs assemble each image's tarball in the workspace. Zero affected
packages or images means zero legs: a legless matrix job vanishes instead
of running, and an empty selection writes no DAG file at all. Forgejo
queues jobs like any task and its runner capacity is the concurrency
control — nothing here hardcodes a parallelism number.

**Pull requests** run the same matrices secret-free: package legs build
with a throwaway key and resolve their affected closure locally
(`single-package.py` local mode), image legs build each image's local
closure and assemble it with `single-image.sh`; tarballs stay in the
ephemeral workspace (no artifact service).

**Pushes to main and manual dispatches** publish: DAG package jobs sign
with the organisational `APK_SIGNING_KEY`, upload to the Forgejo Alpine
repository `v3.24/main` and round-trip anonymously (publication plus
verified round-trip is what releases a job's dependents); DAG image jobs
`apko publish` from that repository and `oras cp`-tag `latest` in the
same job. The publish gate is the event: only pushes to main and manual
dispatches generate the DAG; dispatching ci.yml with `base` (empty
selects everything) and `scope` (all/packages/images) IS the republish
path: bootstraps and recovery without a commit touching every recipe. The version bump check skips
itself when there is no base; recovery caveat: rebuilding in place with
unchanged versions lets a dependent leg resolve the previous build — bump
epochs when rebuilding because content was wrong.

Forgejo signs the published Alpine repository index itself; apko trusts its
live repository key, not the key any individual build signed with. Local
melange build/test guests instead trust the key used to sign their local
package repository (or, in remote mode, additionally the live Forgejo
repository key).

## Required secrets

- `APK_SIGNING_KEY` — organisational secret; the RSA private key (PEM)
  melange signs release packages with. Referenced only by the key step in
  the generated DAG's package jobs; the public key is derived with
  `openssl` for the build/test guests.
- `RELEASE_PACKAGE_TOKEN` — Forgejo token with package write scope
  (package upload) and repository write scope (code and branches). The
  publish step in the generated DAG's package jobs uses it for the
  upload; the plan job uses it to push the `ci/dag-<run_id>` branch that
  carries the generated publish workflow, and the DAG's cleanup job uses
  it to delete the branch again. No runs API polling anywhere — the push
  event is the dispatch. Registry pushes do not use it: they authenticate
  with a Forgejo Authorized Integration ID token — the generated workflow
  sets `enable-openid-connect: true` (exposing the runner's OIDC endpoint)
  and `docker-login` (audience `u:1:ea30cbb9-71af-4074-9cf7-603f4e77208f`)
  in the image jobs mints a JWT for it and logs in to this instance's
  registry with it.

The former `RELEASE_DAG_TOKEN` (second token of the previous two-token
split), `RELEASE_MIRROR_TOKEN` (dispatch mirroring) and
`RELEASE_COORD_TOKEN` (dispatch coordination of the pre-matrix fan-out)
are no longer referenced and can be deleted server-side.

## Checks

```sh
tooling/all-packages.py       # all local packages in dependency order -> dist/
tooling/all-packages.py --dry-run  # print the order without building
tooling/all-images.py         # all images from dist/ packages -> dist/images/
tooling/single-package.py miniflux   # local package build, same flags as CI -> dist/
tooling/single-image.sh miniflux     # local image build from dist/ packages -> dist/images/
tooling/single-test.sh miniflux    # run the recipe's `test:` pipeline via melange test
tooling/all-tests.py         # test every recipe that has tests, from dist/ packages
tooling/tools.sh              # rebuild the tools themselves
tooling/affected.py --base <ref>   # packages/images a change range affects (CI plan)
                                   #   --lists-dir: loop inputs for the CI jobs
tooling/test-release-workflows.py  # offline checks of workflow isolation and gating
tooling/test-affected.py      # offline checks of the selection logic
tooling/test-build-images.sh       # loaded builder images: offline multistage builds
# Optional nested Buildah check, in a disposable privileged container:
tooling/test-build-images.sh --privileged-runner
```

The build-all wrappers require [uv](https://docs.astral.sh/uv/getting-started/installation/),
which manages Python and the inline-pinned PyYAML dependency automatically—no
manual virtualenv or system Python packages. `all-packages.py` discovers
`packages/*.yaml`, orders local build dependencies and their runtime
dependencies (including named subpackages), and calls the signed local builder
once per recipe. `os-release` is prioritized, cycles are rejected before
building, and the first failed build stops the run. `all-images.py` builds
every image definition from the packages already in `dist/`. External Alpine
dependencies are resolved by melange as usual.

For a newly migrated image, build the shared OS metadata package and its
application package before assembling the image, for example:

```sh
tooling/single-package.py os-release
tooling/single-package.py golink      # auto-builds go-licenses first (recipe-planned dependency)
tooling/single-image.sh golink
```

`packages/go-licenses.yaml` pins and builds the license collector from source.
Golink installs that APK in its build environment and runs `go-licenses save`;
it no longer compiles the collector via `go run` during each application build.
Other Go recipes declare the same build dependency where needed; the planner
reads those declarations rather than maintaining a service list. For local builds, rebuild
the tool when its recipe changes before building its consumers.

CI publishes affected packages before affected images in the same run; the
manual sequence below is only needed when iterating locally.

Local builds sign with a gitignored throwaway keypair (`keys/local.rsa`),
mirroring production's signed-index trust path; load the result with
`docker load < dist/images/<name>.tar`.

After building a package, run its recipe's `test:` pipeline with
`tooling/single-test.sh <name>`, or all of them with `tooling/all-tests.py`
(see the Checks section). The pipelines
exercise the packaged binary in a melange test guest: file layout and
notices, startup behavior, and service probes. Checks that cannot run
inside a test guest — dockercleanup's prune cycle against a live Docker
daemon — were verified manually during the migration and are documented in
[MIGRATION.md](MIGRATION.md).
