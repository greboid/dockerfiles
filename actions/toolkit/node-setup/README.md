# node-setup

Installs Node.js and a package manager (pnpm by default) and adds them to `PATH` for later steps. With no inputs, the version is read from `.nvmrc` or `package.json`.

## Inputs

| Input               | Default          | Description                                                       |
| ------------------- | ---------------- | ----------------------------------------------------------------- |
| `node-version`      |                  | Version to install: `18.19.0`, `18`, `lts/*`, `lts/Jod` or `latest` |
| `node-version-file` |                  | `.nvmrc` or `package.json` to read the version from              |
| `package-manager`   | `pnpm`           | `pnpm`, `yarn` or `npm`                                          |
| `working-dir`       | workspace        | Directory containing the project's `package.json`                 |
| `registry-url`      |                  | npm registry to write to the project `.npmrc`                      |
| `token`             | `GITHUB_TOKEN`   | Auth token for `registry-url`                                     |
| `install-dir`       | `.actions/node`  | Directory to install into                                         |
| `force`             | `false`          | Re-download even if already installed                             |
| `debug`             | `false`          | Enable debug logging                                              |

## Outputs

| Output         | Description                   |
| -------------- | ----------------------------- |
| `node-version` | The installed Node.js version |

## Example

```yaml
- uses: https://git.mouse-lake.ts.net/actions/toolkit/git-checkout@master
- uses: https://git.mouse-lake.ts.net/actions/toolkit/node-setup@master
  with:
    node-version: lts/*
- run: pnpm install --frozen-lockfile
```
