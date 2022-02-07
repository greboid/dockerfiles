#!/bin/sh -l

set -eux

rm $HOME/.docker/config.json || true

go install github.com/csmith/contempt/cmd/contempt@latest
git config user.name "$GIT_USERNAME"
git config user.email "$GIT_EMAIL"
buildah login -u $REG_USER -p $REG_PASS $REGISTRY
buildah login -u $REPO_OWNER -p $CONTEMPT_TOKEN $MIRROR_TARGET
ls "$XDG_RUNTIME_DIR/containers/auth.json"
env
contempt --commit --build --push . .
git push
