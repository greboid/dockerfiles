#!/bin/sh
# corepack isn't packaged for alpine 3.24; install it from npm. It
# reads the pnpm version pinned in package.json's packageManager
# field (pnpm@11.22.0).
npm rm -g corepack >/dev/null 2>&1 || true
npm install -g corepack
corepack enable
pnpm install --frozen-lockfile
