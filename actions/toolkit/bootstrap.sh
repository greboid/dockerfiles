#!/bin/sh
# Build the action images locally and push them to the instance registry —
# the same builds and refs the images workflow (.forgejo/workflows/images.yml)
# produces on master. Use it to bootstrap the docker:// refs in the action.ymls
# or to publish images without waiting for CI.
#
# Usage:
#   ./bootstrap.sh                    # build and push every action image
#   ./bootstrap.sh go-build go-test   # just these actions
#   BUILDER=buildah ./bootstrap.sh    # force a builder (auto: docker, else buildah)
#
# Credentials (registry login is skipped when PASSWORD is empty):
#   REGISTRY=git.mouse-lake.ts.net USERNAME=you PASSWORD=... ./bootstrap.sh
set -eu

REGISTRY=${REGISTRY:-git.mouse-lake.ts.net}
REPOSITORY=${REPOSITORY:-actions/toolkit}
TAG=${TAG:-dev}

# keep in sync with the images workflow and the Makefile images target
ACTIONS='git-checkout go-setup node-setup apko-changelog docker-login docker-build docker-push go-build go-test go-vet go-lint git-commit git-try-push release'

usage() {
	sed -n '2,13p' "$0" | sed 's/^# \{0,1\}//'
	exit 0
}
case ${1:-} in
-h | --help) usage ;;
esac

if [ $# -gt 0 ]; then
	ACTIONS="$*"
fi

if [ -z "${BUILDER:-}" ]; then
	if docker info >/dev/null 2>&1; then
		BUILDER=docker
	else
		BUILDER=buildah
	fi
fi

build() {
	case $BUILDER in
	docker) docker build -f "Dockerfile.$1" -t "$REGISTRY/$REPOSITORY/$1:$TAG" . ;;
	buildah) buildah build -f "Dockerfile.$1" -t "$REGISTRY/$REPOSITORY/$1:$TAG" . ;;
	esac
}

push() {
	case $BUILDER in
	docker) docker push "$REGISTRY/$REPOSITORY/$1:$TAG" ;;
	buildah) buildah push "$REGISTRY/$REPOSITORY/$1:$TAG" ;;
	esac
}

if [ -n "${PASSWORD:-}" ]; then
	case $BUILDER in
	docker) printf '%s' "$PASSWORD" | docker login "$REGISTRY" -u "${USERNAME:?USERNAME required with PASSWORD}" --password-stdin ;;
	buildah) printf '%s' "$PASSWORD" | buildah login "$REGISTRY" -u "${USERNAME:?USERNAME required with PASSWORD}" --password-stdin ;;
	esac
else
	echo "PASSWORD not set; assuming $REGISTRY is already logged in with $BUILDER" >&2
fi

cd "$(dirname "$0")"
for a in $ACTIONS; do
	if [ ! -f "Dockerfile.$a" ]; then
		echo "no Dockerfile.$a; skipping" >&2
		continue
	fi
	echo "== $a: build" >&2
	build "$a"
	echo "== $a: push" >&2
	push "$a"
done
