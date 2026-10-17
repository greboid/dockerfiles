Container image `git.mouse-lake.ts.net/containers/sqlite3` — the SQLite
command line client built with apko from upstream Alpine's `sqlite` package
at an exact pinned version, so version bumps arrive as depbot PRs
(`.depbot.yaml`); there is no matching repo in the `packages` org
(upstream: https://www.sqlite.org, packaged by Alpine in `main/sqlite`).

- Tags: `:YYYY.M.D-alpineX.Y` and `:latest`, pushed on every `master` build
- Builds run on push to `master` via the shared workflow in `containers/build-common`
- Package version pins in `apko.yaml` arrive as depbot PRs (`.depbot.yaml`)

## Local build

The repository key is committed under `keys/`, so a plain apko build works:

```sh
apko build apko.yaml sqlite3:local image.tar
```
