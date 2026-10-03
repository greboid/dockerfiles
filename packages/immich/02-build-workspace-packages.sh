#!/bin/sh
# Since v3.2.0 the server package is named "immich" (server/), and
# web's @immich/sdk plus server's @immich/plugin-sdk are unbuilt
# workspace packages compiled before their consumers — upstream
# server/Dockerfile order: sdk -> plugin-sdk -> immich -> immich-web.
pnpm --filter @immich/sdk build
pnpm --filter @immich/plugin-sdk build
pnpm --filter immich build
pnpm --filter immich-web build
