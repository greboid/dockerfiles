# Container migration

Source reference: `~/projects/containers/*/Containerfile`. All 36 application
images have recipes; the reusable `golang`, `rust`, and `alpine-runner`
build/CI images are migrated too. Node is explicitly out of scope for now. Local migration verification is not evidence of production rollout.

## Recipes added

- `miniflux` (existing)
- `golink`
- `ergo`
- `httpredirect`
- `identd`

Custom binaries and assets are built/installed by melange, with apko consuming
exact application package versions. No Docker build stages are needed for
these images. Alpine still supplies standard build and runtime dependencies;
this is not a from-source rebuild of the entire distribution.

## First-batch verification

Both new recipes passed `melange compile`, `apko show-config`, and the normal
local signed package/image builds (`tooling/single-package.py <name>` followed
by `tooling/single-image.sh <name>`).

- Golink: HTTP smoke test with `--dev-listen :8080`; default-path SQLite
  database created with ownership `65532:65532`. Dependency notices verified
  in the image, with neither Go nor `go-licenses` included. Tailnet enrollment
  was not tested. apko restores ownership of the package-provided home directory.
- Shared `go-licenses` build package: local builds and the release workflow's
  build commands tested in a clean temporary repository. Golink uses the
  pinned APK rather than `go run`. The package launcher resolves `GOROOT`
  explicitly because v1.6.0 otherwise silently skips notices when built with
  `-trimpath`; the Golink recipe checks that notices were produced.
- Ergo: image reports `ergo-2.19.1`; packaged language assets and OCI
  user/entrypoint/default arguments checked. Configured IRC/TLS and persistent
  database integration were not tested.

## Second-batch verification

- `httpredirect` v1.1.0 and `identd` v1.0.4: exact upstream archives fetched,
  SHA-256 pinned, and MIT licenses checked. Both passed patched `melange
  compile --arch x86_64`, `apko show-config`, and normal local signed
  package/image builds after rebuilding `os-release` and `go-licenses`.
- Preserved `/httpredirect` and `/identd` entrypoints, empty default arguments,
  UID/GID `65532:65532`, `SSL_CERT_FILE`, and runtime CA certificates/timezones.
  Both retain `/notices` via the shared pinned `go-licenses` build package;
  the release workflow now builds that dependency for both new services.
- Manual runtime checks during migration (the recipes now carry equivalent
  `melange test` pipelines): nonroot/read-only operation, HTTP path/query
  redirects, all redirect environment overrides, default Ident response,
  and Ident flag/environment overrides. Upstream HTTP behavior is 307 by
  default and 308 when `TEMPREDIRECT` is nonempty, despite its README
  claiming the reverse.
- Image entrypoints/default arguments and bundled license notices (including
  Ident's `envflag` dependency) checked; neither Go nor `go-licenses` is in
  either image. These services have no writable data mounts or bundled
  runtime assets. No environmental build blockers encountered.

## Third-batch verification

- `dockercleanup` (upstream tags `v1.0`; packaged as 1.0.0) and `webhooked`
  v1.1.0: exact upstream archives fetched, SHA-256 pinned, MIT licenses
  checked (upstream file spellings differ: LICENSE vs webhooked's LICENCE).
  Both passed patched `melange compile --arch x86_64`, `apko show-config`,
  and the normal local signed package/image builds after rebuilding
  `os-release` and `go-licenses`.
- Preserved `/dockercleanup` and `/webhooked` entrypoints with empty default
  arguments, UID/GID `65532:65532`, `SSL_CERT_FILE`, and runtime CA
  certificates/timezones. Neither service listens on a port or wants a
  writable mount; dockercleanup's legacy `EXPOSE 8080` was vestigial and is
  not kept. The release workflow now builds `go-licenses` for both, too.
- Notice path subtlety: `go-licenses` records each license under the
  imported package's directory, so go-github's license lands at
  `notices/github.com/google/go-github/v72/github/LICENSE` while docker's
  stays at the module root; envflag v2 spells its file `LICENCE`. The
  recipes `test -s` exactly these paths.
- Runtime checks live in the recipes' `melange test` pipelines and pass for
  all migrated packages: image/packaged file layout (entrypoint links,
  bundled notices, installed licenses, no Go tooling), startup behavior,
  and service probes. For dockercleanup they cover the missing-daemon
  failure; the full prune cycle (stopped containers, dangling images,
  `ALLIMAGES`/`CONTAINERS` overrides) was verified manually against a
  throwaway docker:dind daemon and is not reproducible inside a test guest.
- Not testable here: webhooked's actual GitHub hook CRUD (no credentials;
  a dummy token did confirm API reachability, TLS, and 401 handling), and
  dockercleanup against a production daemon. The pinned Docker client
  defaults to API 1.51 and does not negotiate, so daemons older than that
  need `DOCKER_API_VERSION` (same as the legacy image).

## Fourth-batch verification

- `registryauth` v1.0.20, `goplum` v1.1.0, and `puzzles` v1.0.3: exact
  upstream archives fetched, SHA-256 pinned, MIT licenses checked (upstream
  file spellings differ: LICENSE vs goplum's LICENCE — goplum's module is
  `chameth.com/goplum`, so its notice lands at `chameth.com/goplum/LICENCE`).
  All three passed patched `melange compile --arch x86_64`,
  `apko show-config`, normal local signed package/image builds, and their
  new `melange test` pipelines.
- `registryauth`: ships the legacy `/data` directory; the apko image config
  mutates it to `65532:65532` (apko only auto-fixes home directories, so a
  `paths` entry reproduces the legacy chown). Tests cover the `/v2/` Bearer
  challenge, bcrypt-verified token issuance, and rejection of invalid
  credentials when no scope is requested. Real pushes/pulls through the
  embedded registry were not exercised (needs a registry client against the
  running instance); the auth and challenge paths are the parts this
  package adds over stock distribution.
- `goplum`: tests cover the config-required failure and the mTLS API
  listener (throwaway certificates are generated inside the test and the
  listener is probed on 7586). The legacy image's `EXPOSE 8080` was
  vestigial (the API is gRPC) and is not kept.
- `puzzles`: ships the wordlists at the legacy `/app/wordlists` path; static
  assets and templates are embedded in the binary. Tests cover the 308 root
  redirect, anagram lookups against the wordlists, embedded assets, and the
  404 page.
- `melange test` notes for this repository: the package under test is
  installed into the guest automatically; test environments add
  `alpine-release` (same pin as the build environments, so apko finds
  `/etc/os-release`) plus the few tools a pipeline needs (busybox, curl,
  openssl); steps run with `set -e -o pipefail`, so avoid
  `producer | grep -q` when the producer writes much (SIGPIPE kills the
  step); test guests are built from `test.environment`, and image-level
  settings (run-as, entrypoint, `paths` ownership) remain verified via
  `apko show-config` and manual checks, not melange tests.

## Source pinning switched to git-checkout

All recipes now check out the pinned upstream tag with melange's
`git-checkout` instead of fetching GitHub tarballs: `repository` + `tag` +
`expected-commit`, with no `strip-components` (the checkout lands directly
in the workspace root). Notes from the conversion and full rebuild:

- `expected-commit` is always the peeled commit of the tag (`git ls-remote
  "refs/tags/vX^{}"`). Most of our upstreams use lightweight tags, where the
  ref hash is already the commit; ergo, golink and miniflux use annotated
  tags. miniflux's pinned commit matches the `vars.commit` already stamped
  into the binary, and the git checkout now makes Go's own VCS stamping
  agree (buildvcs picks up the workspace `.git`), so `go version -m`
  reports the pinned revision and `vcs.modified=false` for every package.
- Build environments gain a pinned `git=2.54.0-r0` (git-checkout needs the
  git binary inside the guest); melange would add an unpinned `git` via the
  pipeline's `needs` otherwise.
- Every package was rebuilt (`tooling/all-packages.py`), all 10 test
  pipelines pass (`tooling/all-tests.py`), and all 10 images were rebuilt
  (`tooling/all-images.py`). The only failures were transient Go module
  proxy EOFs during `go-licenses` notice collection, which cleared on
  retry. The binaries grew their VCS stamps; package versions and layouts
  are unchanged.

## Package builds unified on one planner

After the fourth batch, the build tooling was collapsed onto shared Python:

- `tooling/planner.py` holds the recipe dependency graph (extracted from
  `all-packages.py`): local build dependencies are read from each recipe's
  `environment.contents.packages` (e.g. golink lists `go-licenses`, so the
  graph knows go-licenses must exist first). No hand-maintained lists.
- `tooling/single-package.py <name>` builds one recipe: it plans build
  dependencies from the graph and builds missing ones into the same signed
  repository before the target. Locally that is `keys/local.rsa` → `dist/`;
  the release workflow passes `--key keys/ci.rsa --out-dir dist/release`
  with a fresh ephemeral key. Keys are cheap, so every build is signed and
  its test guest verifies signatures; a CI key is trusted by nothing —
  release integrity still comes from the Forgejo round trip and
  digest-pinned images.
- `single-package.sh` was removed; `all-packages.py` and the workflow call
  the Python directly (the workflow installs uv via `apk add uv`). The
  workflow's hardcoded go-licenses case list is gone, and its publish glob
  matches only the requested package so build-deps sharing the repo are
  not published.
- A first cut had an unsigned CI mode; it was removed. melange rewrites the
  out-dir's APKINDEX on every build, and an unsigned index silently poisoned
  the signed local repo for later builds — with signing being trivial there
  is no unsigned path at all.
- Transient Go module proxy EOFs during go-licenses notice collection are a
  local environmental issue, not a tooling one: there is no retry logic in
  recipes or tooling; just rerun the build.

## Fifth-batch verification

- `centauri` v2.9.0, `centauri-docker-confd` v1.6.2, and `cocoon` v0.11.3:
  lightweight upstream tags pinned by `expected-commit` via git-checkout;
  MIT licences (csmith spells it LICENCE, cocoon LICENSE). All three passed
  `melange compile`, signed local builds, their `melange test` pipelines,
  and image builds.
- `centauri`: the legacy `-tags lego_httpreq` build tag is kept (`go/build`
  `tags:` input), and the six go-licenses ignores for lego's tencent/OCI
  transitive dependencies carry over; notices assert
  `notices/github.com/csmith/centauri/LICENCE`. The legacy VOLUME /data is
  shipped and chowned 65532:65532 via an apko `paths` entry (verified in
  the image layer); entrypoint `/centauri`, no default arguments. Tests
  cover layout and offline config validation (`-validate` on good and bad
  configs). Deliberately no boot probe: centauri's certificate path reaches
  ACME infrastructure — even the selfsigned provider performs ARI updates —
  and a test guest must not trigger network side effects.
- `centauri-docker-confd`: root main package, no license ignores, notices
  at `github.com/csmith/centauri-docker-confd/LICENCE`; entrypoint only, no
  CMD, no port. The test asserts a loud failure without a Docker daemon;
  the real container-watch loop needs a live daemon (environmental).
- `cocoon`: version ldflag `-X main.Version=v<version>` preserved and
  asserted via `--version`; the four go-licenses ignores carry over.
  Upstream's sqlite driver is a cgo stub under CGO_ENABLED=0 — exactly as
  in the legacy image — so real deployments configure
  `COCOON_DB_TYPE=postgres`; this is documented in the image comment and
  the tests stay offline (no run probe: a PDS is a network-facing service).
  Entrypoint `/cocoon` with default argument `run`.

## Sixth-batch verification

- `irc-bot` v5.0.8, `irc-distribution` v1.1.0, `irc-github` v4.0.3,
  `irc-goplum` v4.0.2, `irc-notifier` v1.0.6, and `irc-webhook` v4.1.1:
  lightweight tags pinned by `expected-commit`; MIT licences (greboid
  spells it LICENSE, csmith LICENCE). All six passed `melange compile`,
  signed local builds, their `melange test` pipelines, and image builds.
  `irc-notifier`'s go.mod declares a domain-less module name, so its notice
  path is flat-ish.
- Notice paths: go-licenses records a module's license under the deepest
  imported package's directory, so the plugins' own notices sit at
  `.../<module>/cmd/<name>/LICENSE` (the `test -s` assertions pin exactly
  these); irc-bot's lands at the module root.
- Tests assert offline startup failures only — no IRC or webhook traffic:
  irc-bot fails with "Server is mandatory" before dialing; the plugins fail
  with "plugin RPC token must be set" (the plugin helper validates config
  before creating any connection); irc-webhook additionally requires its
  token database, whose default `DB_PATH=/data/db` fails without a mounted
  /data exactly as in the legacy image, and with `DB_PATH` set to a file it
  reaches the same token check as its siblings.
- The legacy image's `EXPOSE 8080` on irc-bot matched neither listener
  (web 8000, plugin gRPC 8001) and is not kept; the five plugins have no
  listeners of their own (they register over the bot's gRPC).
- Environmental: the bubblewrap runner leaves a stale guest directory under
  /tmp when a build fails, and Go's module cache marks those files
  read-only, so repeated failures filled the /tmp tmpfs (ENOSPC mid-build).
  Cleanup is `chmod -R u+w /tmp/bubblewrap-guest-* && rm -rf
  /tmp/bubblewrap-guest-*`; recorded here as local environment maintenance,
  not a tooling problem.

## Seventh-batch verification

- `dsp` v1.2.6, `githubmirror` v1.0.3, `pdsps` v1.0.0, `thp` v1.2.2,
  `tsp` v1.2.0, and `tsv` v1.2.0: lightweight tags pinned by
  `expected-commit`; MIT licences (dsp/githubmirror LICENSE; the rest
  LICENCE; pdsps's module is `chameth.com/pdsps`). All six passed
  `melange compile`, signed local builds, their `melange test` pipelines,
  and image builds. Notice paths verified empirically: dsp,
  githubmirror (module root LICENSE), pdsps (`chameth.com/pdsps/LICENCE`),
  thp, tsv (module-root LICENCE).
- `thp` and `tsp` keep the legacy `GOEXPERIMENT=jsonv2` via the `go/build`
  `experiments:` input. The legacy `GOTOOLCHAIN=go1.26.3` pin was
  deliberately dropped: it existed for the older builder image, while the
  Alpine go 1.26.8 build environment already exceeds it (a pin would force
  a toolchain download). `tsp`'s legacy build collected no dependency
  notices, so none are packaged; the licence file is still installed per
  repository convention.
- Tests assert offline startup failures only — no tailscale, Docker or
  GitHub contact: dsp exits *cleanly* (exit 0) logging "socket does not
  exist" when the Docker socket is absent (upstream behavior);
  githubmirror prints its usage and exits cleanly without AUTHTOKEN;
  pdsps, thp, tsp and tsv exit non-zero on their missing required
  configuration (upstream URL / WireGuard keys) before any connection.
- Images: `tsp` ships the legacy VOLUME directories `/config` and
  `/home/nonroot/.config`, owned 65532:65532 via apko `paths` (verified in
  the layer); `thp` and `tsv` ship `/config`; `tsv`'s legacy entrypoint
  argument `--tailscale-config-dir=/config` is expressed as apko's
  top-level `cmd:` (the entrypoint type has no args field).

## Eighth-batch verification

- `linx-server` v3.4.1, `purser` v1.2.0, and `soju` v0.10.1: pinned by
  `expected-commit` (soju's tag is annotated — the peeled commit is
  pinned). All three passed `melange compile`, signed local builds, their
  `melange test` pipelines, and image builds.
- `linx-server`: the fork keeps the upstream module path
  `andreimarcu/linx-server`, so the notice assertion is
  `.../andreimarcu/linx-server/LICENSE.txt` (file spelled LICENSE.txt).
  The legacy entrypoint arguments and default arguments are expressed as
  apko's entrypoint command (split into argv) plus top-level `cmd:`.
  `/data/files` and `/data/meta` are shipped and chowned 65532:65532. Its
  test boots the server on loopback with local storage and asserts the
  index answers 200 — fully offline (embedded templates, no external
  calls).
- `purser`: jsonv2 via the `experiments` input, and the eight legacy
  go-licenses ignores carried over. A wrinkle surfaced: the main module
  imports `encoding/json/v2`, so go-licenses' internal `go list` also
  needs `GOEXPERIMENT=jsonv2` — the legacy Dockerfile got this for free
  because the env var persisted across RUN layers. Without it the save
  fails with "build constraints exclude all Go files". `/data/cache` and
  `/data/output` shipped. The test asserts the missing-Docker-daemon
  failure ("Failed to scan containers"), which happens before any trivy
  vulnerability database is downloaded.
- `soju`: dual binaries (soju + sojuctl) built with the `moderncsqlite`
  tag (pure-Go SQLite under CGO_ENABLED=0), and the notice collection runs
  with the same tag via GOFLAGS. The three legacy ignores carry over —
  go-licenses cannot classify soju's own module ("FORBIDDEN") or
  modernc.org/mathutil ("unknown"), so the assertions pin dependency
  notices (go-scfg, xxhash) instead of the main module. `/data` shipped.
  The test asserts the abort on a missing `-config` file, before any
  listener or IRC connection.
- Tooling note: a go-licenses save whose save_path lives inside the
  scanned module tree (i.e. under melange-out) can rediscover its own
  output and recurse until paths exceed the length limit — linx-server hit
  this. Its recipe saves to /tmp inside the guest and copies into the
  destdir; the other recipes are unaffected.

## Ninth-batch verification

- `forgejo` v16.0.5, `forgejo-runner` v13.2.0, and `tailscale` v1.102.5:
  pinned by `expected-commit` (tailscale's tag is annotated — the peeled
  commit is pinned). All three passed `melange compile`, signed local
  builds, their `melange test` pipelines, and image builds.
- `forgejo`: the CGO/sqlite build is preserved (`CGO_ENABLED=1`, tags
  `bindata sqlite sqlite_unlock_notify`), with the frontend built by the
  pinned Alpine nodejs/npm via `make frontend` and assets embedded via
  `make generate-go`. The redundant `-a` rebuild flag from the legacy
  Dockerfile was dropped (fresh guest toolchain anyway). go-licenses
  carries the five legacy ignores and prunes the vendored frontend notice
  trees. The git/bash/rsync runtime tooling the legacy image hand-copied
  into the rootfs now comes from pinned Alpine packages in the image
  definition (verified present in the layers). `/data` shipped; tests
  assert `--version`.
- `forgejo-runner`: built via its own Makefile (`make build VERSION=...`,
  which stamps `ver.version` and adds the static/netgo/osusergo flags).
  buildah now comes from Alpine's pinned package (1.44.0) instead of the
  legacy source build (1.45.1) — a deliberate deviation, one minor version
  behind, per the repository philosophy of using standard Alpine runtime
  dependencies. Because buildah depends on containers-common (which owns
  `/etc/containers/storage.conf`), the legacy overlay-over-fuse storage
  configuration ships at `/usr/share/forgejo-runner/storage.conf` and the
  image selects it with `CONTAINERS_STORAGE_CONF`. fuse-overlayfs and
  gnupg come from Alpine pins. The image sets apko's `work-dir: /data`
  (the legacy WORKDIR). Tests assert `--version`.
- `tailscale`: three binaries (tailscaled + tailscale with version stamps,
  containerboot without, matching legacy), installed at the legacy paths
  (`/usr/local/bin`, `/containerboot`). The four legacy go-licenses ignores
  carry over (freetype raster/truetype are unknown-license modules).
  `/var/run/tailscale` is no longer packaged: melange's tempdir linter
  rightly rejects var/run content, and containerboot creates the directory
  itself at startup (verified in upstream source) before symlinking the
  socket. `/var/lib/tailscale` shipped. Tests assert `tailscale version`.
- go-licenses recursion (from the eighth batch) also applied to
  forgejo-runner and forgejo: both save notices to /tmp inside the guest
  and copy them into the destdir.

## Tenth-batch verification

- `postgres-18` (18.6 + pgvector v0.8.6), `redis` (8.10.2), and `sws`
  (static-web-server v2.44.0): all three passed `melange compile`, signed
  local builds, their `melange test` pipelines, and image builds.
  postgres uses `fetch` (the tarball is upstream's signed release artifact
  with the legacy SHA-256 pin); redis and sws use `git-checkout` (sws's
  tag is annotated — the peeled commit is pinned).
- `postgres-18`: configure/make world-bin into /usr/local/pgsql. The
  legacy `entry.sh` init script ships verbatim (tabs widened; heredoc
  terminators moved to column 0 because `<<-` only strips tabs) at
  `/usr/local/bin/entry.sh`. Build fixes found on the way: PG 18's
  configure demands bison/flex even from a tarball, pg_combinebackup needs
  linux-headers (linux/fs.h), and Alpine 3.24 renamed util-linux-libs to
  libuuid. `/var/lib/postgresql/data` (0o750, 65532) and `/tmp` (1777)
  shipped; the legacy `/var/run` directory was dropped (the Unix socket
  defaults to /tmp). The image reproduces the legacy env (PATH prepend,
  LANG, PGDATA/PGHOST/PGUSER) and STOPSIGNAL SIGINT. The test initializes
  a real cluster as the guest's build user (postgres refuses root) and
  runs a socket-only query — fully offline.
- `pgvector`: split out of the postgres-18 package into its own recipe
  (upstream v0.8.6) so it can be bumped independently. It declares
  postgres-18 as a build and runtime dependency, builds with
  `OPTFLAGS=""` (drops -march=native), and stages via PGXS `DESTDIR` into
  its own tree (`lib/vector.so`, `share/extension/*`, headers) — the image
  combines the two packages and owns nothing of each other's files. Its
  test boots a cluster from the postgres-18 dependency and asserts
  `CREATE EXTENSION vector` reports 0.8.6 — fully offline.
- `redis`: static build (LDFLAGS=-static, bundled jemalloc/openssl/zlib
  static libs), binary at `/redis` with `/redis.conf` at the root, exactly
  as legacy. `/home/nonroot/database` shipped (the config's `dir`). The
  test starts the server and asserts a RESP `+PONG` over nc. Redis 8 is
  tri-licensed (RSAL-2.0 OR SSPL-1.0 OR AGPL-3.0-only).
- `sws`: cargo release build of static-web-server with the local
  query-preserving-rewrite patch applied (embedded byte-exact in the
  recipe; note unified-diff context lines' leading spaces must survive
  YAML transcription). Dual-licensed Apache-2.0 OR MIT. The test boots the
  server against a temporary document root and asserts a 200.

## Eleventh-batch verification

- `immich` v3.2.0 and `immich-ml` v3.2.0 (both AGPL-3.0-only, one monorepo
  tag pinned by `expected-commit`; libvips split into its own package
  v8.18.6, LGPL-2.1-or-later): all passed `melange compile`, signed local
  builds, their `melange test` pipelines, and image builds.
- `libvips`: built with auto-features disabled and only immich's loaders
  (heif linked in; codec plugins libheif-dav1d/libde265 arrive via libheif,
  jpeg/openjpeg pinned explicitly). Headers ship in the main package under
  the `vips/` namespace that sharp's `<vips/vips8>` expects; the internal
  `memory.h`/`semaphore.h` are dropped (musl-dev collisions — the legacy
  dodged them via the isolated /usr/local prefix). rsvg is deliberately
  disabled: the librsvg-dev chain trips an apk world-solver pc: conflict
  under melange, and immich rejects SVG uploads anyway.
- `immich`: pnpm monorepo build following upstream's order (corepack from
  npm, pnpm@11.22.0 via packageManager; sdk → plugin-sdk → immich →
  immich-web), production deploy with --prod --no-optional, sharp
  source-built against the system libvips (SHARP_FORCE_GLOBAL_LIBVIPS),
  cross-platform prebuild cleanup, geodata fetched with pinned SHA-256
  hashes (geonames dumps drift; the date stamp derives from
  SOURCE_DATE_EPOCH). Runtime dependencies are enumerated explicitly:
  nodejs, jellyfin-ffmpeg (tonemapx), perl + coreutils-env (exiftool's
  env shebang), coreutils (fluent-ffmpeg's hardcoded `nice`),
  icu-data-full (node aborts without full ICU data), libvips + the heif
  codec plugins. Tests are offline end-to-end smokes: HEIC decode through
  sharp→libvips→libheif→libde265, a nice-resolved fluent-ffmpeg
  transcode, and the jellyfin `tonemapx` HDR filter. The image sets the
  legacy env (NODE_ENV/IMMICH_*/HOME, LIBC=musl), work-dir
  /usr/src/app/server, entrypoint `node dist/main`, upload volume 65532.
- `immich-ml`: venv over Alpine's ML stack (--system-site-packages; the
  six heavyweight deps come from pinned py3-* packages, everything else
  from the frozen uv.lock export with onnx/ml-dtypes source-built). The
  legacy scanelf closure loops are replaced by apk dependency resolution.
  Runtime deps pinned: python3, the py3-* stack, mimalloc2 (image
  LD_PRELOAD), protobuf (onnx). Test boots the server offline and asserts
  /ping answers 200. Image: full legacy env block (DEVICE=cpu, cache
  folders, LD_PRELOAD), work-dir /usr/src, entrypoint
  `python -m immich_ml`, /cache volume 65532.
- Debugging notes: sharp's binding.gyp + node's common.gypi add
  `-flto=4 -ffat-lto-objects` to addon builds, and gcc15 LTO + Alpine's
  fortify headers fail the sharp link (vsnprintf always_inline) — LTO is
  stripped from the generated makefile between configure and build.
  busybox wget in the guest has no TLS; geodata downloads use curl.

## Remaining application images

- Go services: none — all migrated.
- Multi-binary/integration-heavy Go services: none — all migrated.
- Native/Rust services: none — all migrated.
- Node/Python services and native dependencies: none — all migrated.

All application images from the legacy monorepo are now migrated. The legacy
build-infrastructure review is recorded below. The user confirmed external
consumers of `golang` and `rust`, and continued CI use of `alpine-runner`:
all three are now migrated with legacy versions and interfaces preserved.
Node is explicitly excluded for now; the user will recreate it if needed.

Preserve build tags, version flags, bundled data, entrypoint/default argument
boundaries, configuration paths, environment variables, and writable mounts.
In particular, inspect the `jsonv2` builds, Soju's database build tags,
Forgejo's Git dependency, and Tailscale's networking tools/capabilities before
porting them. Package additional custom components separately where shared.

## Legacy build infrastructure

The initial review considered only consumers visible in this repository and
incorrectly concluded that replacing application build dependencies removed
the need for reusable builder images. The user confirmed that `golang`,
`rust`, and `alpine-runner` remain required. Migrate them with their legacy
versions, paths, environment, and tools intact; do not substitute Alpine's
older Go/Rust/Buildah versions. Check with the user before retiring any other
infrastructure image based solely on local reference searches.

| Legacy component | Replacement / disposition |
| --- | --- |
| `alpine` | Melange guests and apko images resolve Alpine packages directly; no minirootfs image is needed here. Keep the legacy image while its 42 Containerfile consumers remain active. Its SHA-256-verified minirootfs bootstrap must not be replaced with an unchecked download. |
| `baseroot`, `base` | apko supplies filesystem/account setup and resolves runtime libraries; each image declares CA certificates, timezones, `SSL_CERT_FILE`, and its runtime user. Writable paths are explicit. The legacy `/gitconfig` only suppressed detached-HEAD advice and was not a runtime service. No standalone base image is needed by these recipes. |
| `golang` | Migrated (twelfth batch): source-built Go 1.27.1 at `/usr/local/go`, go-licenses v2.0.1 with its toolchain-aware launcher, writable GOPATH, legacy env and replaceable shell CMD. Existing application recipes' Alpine Go and go-licenses v1.6.0 are unchanged. |
| `rust` | Migrated (twelfth batch): rustup 1.29.1 and Rust 1.98.1 musl toolchain, checksum-pinned upstream binaries with restored upstream manifest state, writable homes, fixed PATH. The sws recipe's Alpine compiler is unchanged. |
| `git` | The legacy scratch image was a file donor for forgejo and forgejo-runner, not a service. Both apko images install Alpine git 2.54.0 instead of the legacy custom static Git 2.55.0/curl build. Build guests also declare Git directly. |
| `alpine-runner` | Migrated (twelfth batch): CI job image with source-built Buildah 1.45.1 over the shared golang-builder, legacy fuse-overlayfs storage configuration via `CONTAINERS_STORAGE_CONF`, root shell with apk, declared CI tools. This is not the `forgejo-runner` daemon image; host privileges and the runner-label mapping still need deployment verification before switching the runner label. |
| `build-common` | Replaced here by the three local workflows and shared tooling/planner. Retain the old reusable workflows until their callers are switched off: the legacy checkout still has 86 calls across 43 repositories. |

### Release-workflow verification after the review

- Fixed step-local environment leaks: package publish and anonymous
  round-trip steps now each receive `PACKAGE`; the image `latest` step
  restores `DOCKER_CONFIG` to the publish step's credentials directory.
- `tooling/test-release-workflows.py` executes those actual YAML shell steps
  in fresh processes with fake curl/apko/oras commands. Both checks pass,
  and both fail against the pre-fix workflows. It checks that build-only
  dependencies are not uploaded and that latest is copied from the published
  digest using the existing auth file. PR validation runs these checks too.
  No real registry writes are made.
- Corrected stale unsigned-package documentation: local/CI build repositories
  are signed for melange guests; the CI key is ephemeral. Published-package
  integrity uses the anonymous byte-for-byte HTTPS round trip and Forgejo's
  signed index. This is not OCI image signing.
- Both planner dry runs pass (40 package recipes, 36 images). No application
  recipes changed in this review, so application builds/tests were not rerun.

### Remaining operational cutover

1. Inspect the registered `alpine` runner label on the actual Forgejo runner.
   All three workflows use it, but neither repository establishes which
   image/host that label selects. Verify checkout's action runtime, Git/Go
   for `tooling/tools.sh`, curl, permission to install uv via apk, and a
   Docker-compatible melange `--runner docker` environment. Buildah plus
   fuse-overlayfs alone does not establish Docker runner compatibility.
   Migrate the existing job image's declared interface now; verify its
   deployment against the actual host's socket/privilege arrangement before
   switching the runner label.
2. Run one real package release (including a local build dependency) and
   image release, then verify the Forgejo round trip, commit tag,
   `latest`, and optional mirror. Offline workflow checks do not validate
   runner privileges, registry credentials, or Forgejo's execution semantics.
3. Coordinate caller/deployment cutover before disabling the old repositories.
   New releases are manual dispatches, whereas the legacy repositories invoke
   shared workflows on push/PR. The legacy build-common uses Authorized
   Integration OIDC; the new workflows use scoped package-write secrets.
   Confirm that credential and trigger change operationally; do not silently
   inherit old toolkit actions or copy credentials into image recipes.
4. Retire old builders/base images only after checking external consumers and
   disabling the legacy publishing jobs. Otherwise both pipelines can still
   publish the same application's `latest` tag. No old workflows, registry
   images, or deployments were modified during this review.

## Twelfth batch: reusable builder and CI images

The user confirmed external consumers of `golang`, `rust`, and
`alpine-runner`: these images are migrated as reusable build/CI images with
their legacy versions, paths, environment and tools intact. Node is
explicitly out of scope (the user will recreate it if needed). All three
images keep the legacy parent's interface: default root, replaceable shell
CMD (no entrypoint), `LANG=en_US.UTF-8`, working apk with trusted Alpine
keys and CDN repositories, and the repository-standard nonroot 65532
declared so `/etc/passwd` and `/etc/group` are generated (build stages may
drop to nonroot).

- `golang` (Go 1.27.1): `golang-builder` source-builds the pinned Go
  tarball (legacy SHA-256) into `/usr/local/go` using Alpine go as the
  bootstrap, prunes the same payload directories as the legacy image, and
  ships the writable `/go` tree; `go-licenses-builder` pins the legacy
  collector v2.0.1 separately from the applications' v1 and keeps the
  GOROOT-resolving launcher. Image env: `GOPATH=/go`,
  `GOFLAGS=-buildvcs=false`, `GOTELEMETRY=off`, `GOAMD64=v1`,
  `SOURCE_DATE_EPOCH=0`, legacy PATH, and the inherited
  `[advice] detachedHead = false` gitconfig. Verified: `go version`
  reports 1.27.1, pure-Go and cgo builds compile in the package test and
  again inside the loaded image, the v2 collector reports stdlib-only
  modules and handles a relocated GOROOT, `apk add` works in the image,
  and a foreign multi-stage Dockerfile builds against it.
- `rust` (Rust 1.98.1, rustup 1.29.1): a documented compatibility
  exception packages the exact rustup-distributed musl toolchain (a source
  bootstrap would not preserve these artifacts): rustup-init, the release
  manifest and all three components are SHA-256 pinned, installed from an
  offline local mirror, then the stored manifest URLs/hash are restored to
  upstream so future rustup operations keep working. Legacy
  `RUSTUP_HOME`/`CARGO_HOME` (world-writable, as the legacy image left
  them — the `worldwrite` linter is deliberately disabled on the recipe)
  and the cargo symlink ship in the package. The legacy PATH contained a
  semicolon typo that hid `cargo/bin`; the image fixes it to a colon
  explicitly. Verified: rustc/cargo/rustup report the pinned versions,
  rustup state (minimal profile, musl host, components) is intact, real
  rustc and cargo musl compilations run offline in the package test, and
  a foreign multi-stage Dockerfile builds a cargo project against the
  loaded image.
- `alpine-runner` (CI job image): `buildah` source-builds the legacy
  custom v1.45.1 (`make bin/buildah`, `CGO_ENABLED=1`) with
  `golang-builder` as the build dependency and the default upstream tag
  set (seccomp, btrfs driver via btrfs-progs-dev, no gpgme tag — but the
  gpgme C library is still compiled into containers/image's default path,
  so the build guest needs `gpgme-dev` and the image ships `gpgme`, as
  the legacy stages did). The recipe deliberately reuses Alpine's package
  name and wins on version over Alpine's 1.44.0-r1. `alpine-runner`
  (package) ships the verbatim legacy storage.conf at the private
  `/usr/share/alpine-runner/storage.conf` (selected via
  `CONTAINERS_STORAGE_CONF`, as in forgejo-runner) plus the root
  passwd/group entries. Declared CI tools (git, curl, wget, openssl, the
  gnupg stack, fuse-overlayfs, netavark) come from pinned Alpine packages.
  Verified: `buildah version` reports 1.45.1 in the loaded image, and in
  a privileged container the store initializes through fuse-overlayfs and
  a full scratch build/from/mount lifecycle succeeds.
- Integration findings (leader-side full builds, all passing with their
  `melange test` pipelines): melange resets the guest PATH regardless of
  `environment.environment`, so pipeline steps export their toolchain
  paths explicitly; build guests need `ca-certificates-bundle` (plain
  `ca-certificates` failed TLS in `fetch` steps); busybox tar cannot parse
  GNU-format archives (rust's license extraction uses Alpine GNU tar);
  the toolchain recipes carry foreign-arch ELF in upstream sources, so
  `golang-builder` sets `no-depends`/`no-provides` with explicit runtime
  deps; and `tooling/single-image.sh` now feeds the local repository via
  apko's build-only `-b` flag so `/etc/apk/repositories` in reusable
  images only ever names Alpine mirrors.
- apko account behavior (verified empirically): `accounts.users`
  generates `/etc/passwd`/`/etc/group` containing only declared users, so
  the images declare the standard nonroot 65532; entries needed by legacy
  binaries (root, for buildah's user resolution) are shipped by the
  package and merged by apko. `/var/tmp` (buildah's staging area in the
  legacy minirootfs) is an explicit `paths` entry in alpine-runner.

## Verification for each batch

1. Fetch the exact upstream release, verify its checksum and license, and
   inspect its build and runtime requirements.
2. Validate melange and apko recipes using the repository's patched tools.
3. Build each application APK with `tooling/single-package.py` (it plans
   and builds any local build dependencies first), then its apko image
   using `tooling/single-image.sh`.
4. Add a `test:` pipeline to the recipe and run it with
   `tooling/single-test.sh <name>`: packaged file layout, startup
   behavior, and service probes where the guest allows. Record anything
   the guest cannot reproduce (live daemons, real credentials) as a manual
   verification or an environmental limitation rather than treating recipe
   validation as a successful build.

