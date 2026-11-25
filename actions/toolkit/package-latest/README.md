# package-latest

Looks up the latest version of a package in the Forgejo/Gitea instance's
package registry and reports the next version to release. It lists the
owner's packages (`GET /api/v1/packages/{owner}`, page by page), filters by
package type, package name and an optional version prefix, and picks the
highest version. Versions that are not semver (the evergreen `latest` tag of
a container package, say) are ignored.

Outputs:

- `version` — the latest matching version, empty when the registry has none
- `next` — the version a new release should use: `version` with its patch
  component incremented (`1.20261007.3` → `1.20261007.4`), or the prefix with
  a `0` appended when nothing matched (`1.20261007.` → `1.20261007.0`), or
  `0.0.0` without a prefix

This replaces the curl/jq pipelines scheduled image builds otherwise need to
compute the day's build number.

A small static Go binary in an Alpine image (docker action) — no Node.js, no
`actions/node20` runtime, no jq. Public packages need no credentials; the
`token` input authenticates private ones.

## Usage

```yaml
steps:
  - uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@master
  - name: Set date in var
    run: echo "VERSTRING=$(date -u +%Y%m%d)" >> $GITHUB_ENV
  - name: Determine image version
    id: pkg
    uses: https://git.mouse-lake.ts.net/actions/toolkit/package-latest@master
    with:
      type: container
      package: dockerbase/root
      prefix: 1.${{ env.VERSTRING }}.
  - name: Use it
    run: echo "IMAGE_TAG=${{ steps.pkg.outputs.next }}" >> "$GITHUB_ENV"
```

## Inputs

- `server-url` — instance base URL to query; defaults to the workflow's server URL
- `owner` — owner (user or org) whose packages to list; defaults to the workflow's repository owner
- `type` — package type (`container`, `generic`, ...); empty lists every type
- `package` — package name to match (substring), e.g. `dockerbase/root` for a container package
- `prefix` — only versions starting with this prefix count, e.g. `1.20261007.` (include the trailing separator)
- `token` — access token for private packages; public packages need none

## Outputs

- `version` — the latest matching version, empty when the registry has none
- `next` — the version a new release should use (see above)

## Development

```sh
make check   # go test + go vet (from the repository root)
make images  # build the action image locally, as the images workflow does (from the repository root)
```
