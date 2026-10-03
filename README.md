# forgejo sync monorepo
#
# This repository is a view-only monorepo: an automatically generated snapshot
# of the [`containers`](https://git.mouse-lake.ts.net/containers) and
# [`packages`](https://git.mouse-lake.ts.net/packages) Forgejo organisations.
#
# It is generated output. Do not open pull requests or file issues here, and
# do not commit directly: every change belongs in the organisation
# repositories and will appear here on the next sync.

## Layout

Each organisation repository is snapshotted at the tip of its default branch:

```
containers/<repo>/...   # one repository per container image (apko builds)
packages/<repo>/...     # one repository per package (melange builds)
```

The full source of each repository is present and browsable, but the
monorepo keeps no per-repository history: each sync squashes the current
state of every repository into a single commit.

## Sync

`.forgejo/workflows/sync.yml` runs `sync.sh` hourly on a Forgejo Actions
runner. The script:

1. enumerates the repositories in each organisation via the Forgejo API,
2. shallow-fetches the default branch of every one of them,
3. assembles the tree with `git read-tree --prefix=<org>/<repo>/`,
4. commits and pushes a single `Sync:` commit — only if the tree changed.

Repositories added to an organisation appear after the next sync;
repositories deleted (or emptied) disappear from it.

## Running it manually

```
FORGEJO_TOKEN=... ./sync.sh      # or keep a token in the PAT file (gitignored)
```

The script is safe to re-run: it only pushes when the snapshot differs, and
retries cleanly if the branch moved underneath it.

## History

This repository previously held the original hand-written Dockerfiles for
every image. That content was retired in favour of the per-repository
organisations and removed from this branch; it remains in the git history
before the "Convert to view-only monorepo" commit.
