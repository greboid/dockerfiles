#!/bin/sh
# Build an image locally from images/<name>.yaml, consuming the packages
# built locally by tooling/single-package.py (dist/x86_64/). Build only —
# nothing is pushed anywhere.
#
#   tooling/single-package.py miniflux     # package first, if needed
#   tooling/single-image.sh miniflux
#
# Output: dist/images/<name>.tar (an OCI layout tarball; load with
# docker/podman load). Requires network for the Alpine CDN base packages.

set -eu

BASE=$(cd "$(dirname "$0")/.." && pwd)
cd "$BASE"

if [ "$#" -ne 1 ]; then
	echo "usage: $0 <image-name>" >&2
	exit 2
fi

NAME=$1
case "$NAME" in
	''|*[!a-z0-9-]*)
		echo "invalid image name: $NAME" >&2
		exit 2
		;;
esac

RECIPE="images/$NAME.yaml"
if [ ! -f "$RECIPE" ]; then
	echo "no such image definition: $RECIPE" >&2
	exit 2
fi

LOCAL_REPO=dist
if [ ! -f "$LOCAL_REPO/x86_64/APKINDEX.tar.gz" ]; then
	echo "no local packages in $LOCAL_REPO/x86_64 — run tooling/single-package.py first" >&2
	exit 2
fi

APKO=tooling/bin/apko
if [ ! -x "$APKO" ]; then
	tooling/tools.sh apko
fi

KEY=keys/local.rsa.pub
if [ ! -f "$KEY" ]; then
	echo "missing local signing key $KEY — run tooling/single-package.py first" >&2
	exit 2
fi

mkdir -p dist/images
OUTPUT=dist/images/$NAME.tar

epoch=$(git log -1 --format=%ct)
export SOURCE_DATE_EPOCH="$epoch"

# The local repository is a build-only input: never write a host-relative
# dist/ URL into /etc/apk/repositories in reusable build/CI images.
# Base packages resolve from Alpine and custom packages from dist/.
# apko appends <arch>/APKINDEX.tar.gz to local repo paths itself, and
# verifies the index against our local public key. SBOMs land next to the
# tarball instead of littering the working directory.
exec "$APKO" build \
	-b "$LOCAL_REPO" \
	-k "$KEY" \
	--sbom-path "$(dirname "$OUTPUT")" \
	"$RECIPE" \
	"$NAME:local" \
	"$OUTPUT"
