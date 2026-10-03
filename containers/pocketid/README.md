# pocket-id

Container image `git.mouse-lake.ts.net/containers/pocket-id` — built from source following the fleet pattern.

- Tags: `:YYYYMMDD-alpineX.Y` (immutable, never overwritten) and `:latest`
- Builds run on push to `master` via the shared workflow in `containers/build-common`
- Updates arrive as depbot PRs (`.depbot.yaml`)
- Upstream source: https://github.com/pocket-id/pocket-id

Frontend is built with pnpm and embedded into the static Go binary (`backend/frontend/dist`), version stamped from the upstream `.version` file. Runs as UID 65532 with `/app/data` for the SQLite database and uploads. Requires `APP_URL` and `ENCRYPTION_KEY` (see upstream docs). Port 1411; `/pocket-id healthcheck` hits `/healthz`.
