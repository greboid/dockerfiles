#!/bin/sh -l

set -eux

buildah login -u $REG_USER -p $REG_PASS $REGISTRY
buildah login -u $REPO_OWNER -p $CONTEMPT_TOKEN $MIRROR_TARGET

for IMAGE in $MIRROR_IMAGES; do
  if buildah images $REGISTRY/$IMAGE >/dev/null 2>&1; then
    buildah tag $REGISTRY/$IMAGE $MIRROR_TARGET/$MIRROR_PATH/$IMAGE
    buildah push $MIRROR_TARGET/$MIRROR_PATH/$IMAGE
  fi
done
