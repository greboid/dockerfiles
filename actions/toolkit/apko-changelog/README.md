# apko-changelog

Compares two apko lock files and produces a commit-message style changelog.
For every architecture whose packages differ between the old and the new lock
file it lists each change — version bumps as `package: old => new`, plus
additions and removals — in sorted, deterministic order. Repository changes
are listed separately as non-package changes (for example an Alpine version
bump, which only rewrites repository URLs). The message is exposed as the
`message` step output; `changes` is `true` when any package changed.

Example output:

```
Updating to 20260714 with the following package changes:
 - aarch64
   - ca-certificates-bundle: 20260611-r0 => 20260909-r0
   - tzdata: 2026d-r0 => 2026e-r0
 - x86_64
   - ca-certificates-bundle: 20260611-r0 => 20260909-r0
   - tzdata: 2026d-r0 => 2026e-r0
```

A small static Go binary in an Alpine image (docker action) — no Node.js, no
`actions/node20` runtime. It talks to the runner through the standard
`GITHUB_*` files (`-flag` args from the action inputs, `$GITHUB_OUTPUT`).

## Usage

```yaml
steps:
  - uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@master
  - name: Save the previous lock file
    run: cp base.lock.json old.lock.json
  - uses: https://git.mouse-lake.ts.net/actions/toolkit/changelog@master
    id: changelog
    with:
      new-lockfile: base.lock.json
      old-lockfile: old.lock.json
      version: ${{ env.VERSTRING }}
  - run: echo "${{ steps.changelog.outputs.message }}"
```

## Inputs

- `new-lockfile` — path (relative to the workspace) of the new lock file; default `base.lock.json`
- `old-lockfile` — path of the previous lock file; default `old.lock.json`
- `version` — version string mentioned in the message; defaults to the `VERSTRING` environment variable

## Outputs

- `changes` — `true` when any package (not repository) entry differs between the two lock files
- `message` — the multi-line changelog message

## Development

```sh
make check   # go test + go vet (from the repository root)
make images  # build the action image locally, as the images workflow does (from the repository root)
```
