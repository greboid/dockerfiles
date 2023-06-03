#!/bin/sh -l

set -eux

go install github.com/csmith/contempt/cmd/contempt@latest
git config user.name "$GIT_USERNAME"
git config user.email "$GIT_EMAIL"
contempt -template Containerfile.gotpl -output Containerfile --commit --build --push . .
git push
