#!/bin/sh
# Run the melange test pipeline for one package recipe locally, mirroring
# how CI would run it. The package must already be built into dist/ (see
# tooling/single-package.py); the test guest is assembled from the recipe's
# test.environment plus the package from our local repository.
#
#   tooling/single-package.py registryauth   # build first, if needed
#   tooling/single-test.sh registryauth
#
# Tests live in the `test:` section of each packages/<name>.yaml so the
# checks stay next to the recipe that builds the package. Use
# tooling/all-tests.py to test every recipe that has tests.

set -eu

BASE=$(cd "$(dirname "$0")/.." && pwd)
cd "$BASE"

if [ "$#" -ne 1 ]; then
	echo "usage: $0 <package-name>" >&2
	exit 2
fi

NAME=$1
case "$NAME" in
	''|*[!a-z0-9-]*)
		echo "invalid package name: $NAME" >&2
		exit 2
		;;
esac

RECIPE="packages/$NAME.yaml"
if [ ! -f "$RECIPE" ]; then
	echo "no such recipe: $RECIPE" >&2
	exit 2
fi

MELANGE=tooling/bin/melange
if [ ! -x "$MELANGE" ]; then
	tooling/tools.sh melange
fi

KEY=keys/local.rsa.pub
if [ ! -f "$KEY" ]; then
	echo "missing local signing key $KEY — run tooling/single-package.py first" >&2
	exit 2
fi

if ! ls dist/x86_64/ 2>/dev/null | grep -q "^$NAME-"; then
	echo "no built package in dist/x86_64/ — run tooling/single-package.py $NAME first" >&2
	exit 2
fi

exec "$MELANGE" test "$RECIPE" \
	--arch x86_64 \
	--runner docker \
	--repository-append "$BASE/dist" \
	--keyring-append "$KEY"
