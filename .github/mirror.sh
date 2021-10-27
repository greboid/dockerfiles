#!/bin/sh -l

set -eux

buildah login -u $REGISTRY_USER -p $REGISTRY_PASS $REGISTRY
buildah login -u $REPO_OWNER -p $GITHUB_TOKEN $MIRROR_TARGET

for IMAGE in $MIRROR_IMAGES; do
  if buildah images $REGISTRY/$IMAGE >/dev/null 2>&1; then
    echo $IMAGE changed, mirroring
  else
    echo $IMAGE not changed
  fi
done