#!/bin/sh -l

set -eux

go install github.com/csmith/contempt/cmd/contempt@latest
git config user.name "$GIT_USERNAME"
git config user.email "$GIT_EMAIL"
buildah login -u $REG_USER -p $REG_PASS $REGISTRY
buildah login -u $REPO_OWNER -p $CONTEMPT_TOKEN $MIRROR_TARGET
contempt --commit --build --push . .
git push
