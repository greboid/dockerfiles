#!/bin/sh -l

set -eux

go install github.com/csmith/contempt/cmd/contempt@latest
git config user.name "$GIT_USERNAME"
git config user.email "$GIT_EMAIL"
docker login -u $REG_USER -p $REG_PASS $REGISTRY
docker login $MIRROR_TARGET -u $REPO_OWNER -p $CONTEMPT_TOKEN 
contempt --commit --build --push . .
git push
