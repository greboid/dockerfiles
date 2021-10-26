#!/bin/sh -l

go install github.com/csmith/contempt/cmd/contempt@v1.0.2
git config user.name "$GIT_USERNAME"
git config user.email "$GIT_EMAIL"
buildah login -u $REGISTRY_USER -p $REGISTRY_PASS $REGISTRY
buildah login -u $REPO_OWNER -p $GITHUB_TOKEN $MIRROR_TARGET
contempt --commit --build --push . .
for IMAGE in $MIRROR_IMAGES; do
  buildah images $REGISTRY/$IMAGE >/dev/null 2>&1
  result=$?
  if [ $result -eq 0 ]; then
    buildah tag $REGISTRY/$IMAGE $MIRROR_TARGET/$MIRROR_PATH/$IMAGE
    buildah push $MIRROR_TARGET/$MIRROR_PATH/$IMAGE
  fi
done
git push
