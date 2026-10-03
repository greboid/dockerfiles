#!/bin/sh
# Build our copies of melange, apko and oras from upstream source at the
# versions in versions.env, applying the local patch sets where present.
#
#   ./tools.sh            build all three into bin/
#   ./tools.sh oras       build just one (melange|apko|oras)
#
# Requirements: git, go (any recent toolchain; Go downloads newer ones as
# needed), network access to github.com and the Go module proxy.

set -eu

BASE=$(cd "$(dirname "$0")" && pwd)
cd "$BASE"
# shellcheck source=versions.env
. ./versions.env

OUT=${OUT:-"$BASE/bin"}
mkdir -p "$OUT"

case "${1:-all}" in
	melange) tools=melange ;;
	apko) tools=apko ;;
	oras) tools=oras ;;
	all) tools="melange apko oras" ;;
	*)
		echo "usage: $0 [melange|apko|oras|all]" >&2
		exit 2
		;;
esac

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

for tool in $tools; do
	case "$tool" in
		melange)
			repo=chainguard-dev/melange
			ver=$MELANGE_VERSION
			pkg=.
			;;
		apko)
			repo=chainguard-dev/apko
			ver=$APKO_VERSION
			pkg=.
			;;
		oras)
			repo=oras-project/oras
			ver=$ORAS_VERSION
			pkg=./cmd/oras
			;;
	esac

	echo "== $tool $ver"
	# init + fetch + checkout instead of clone: annotated tags (oras) make
	# shallow clones warn "refs/tags/X is not a commit"; fetching the tag ref
	# directly resolves tag -> commit cleanly.
	git init -q "$TMP/$tool"
	git -C "$TMP/$tool" remote add origin "https://github.com/$repo"
	git -C "$TMP/$tool" fetch -q --depth 1 origin "refs/tags/$ver:refs/tags/$ver"
	git -C "$TMP/$tool" checkout -q --detach "$ver"
	echo "   at $(git -C "$TMP/$tool" rev-parse --short=12 HEAD)"

	(
		cd "$TMP/$tool"
		for p in "$BASE/$tool/patches/"*.patch; do
			[ -e "$p" ] || continue
			git apply "$p"
			echo "   applied $(basename "$p")"
		done
		if [ "$tool" = melange ]; then
			# melange resolves build guests through apko's in-tree apk
			# client, but takes it from the module cache, so the apko
			# patch series never reaches it. Replace the module with a
			# tree at the exact version melange pins, carrying the apko
			# patch that lets remote index reads tolerate servers without
			# HEAD (Forgejo Alpine endpoints 405 the etag probe). Patching
			# the pinned version keeps the module graph and go.sum
			# untouched; a disk replace skips module verification.
			apkover=$(awk '$1 == "chainguard.dev/apko" { print $2 }' go.mod)
			git init -q "$TMP/apko-lib"
			git -C "$TMP/apko-lib" remote add origin "https://github.com/chainguard-dev/apko"
			git -C "$TMP/apko-lib" fetch -q --depth 1 origin "refs/tags/$apkover:refs/tags/$apkover"
			git -C "$TMP/apko-lib" checkout -q --detach "$apkover"
			git -C "$TMP/apko-lib" apply "$BASE/apko/patches/007-index-head-405-fallback.patch"
			go mod edit -replace "chainguard.dev/apko=$TMP/apko-lib"
			echo "   replaced chainguard.dev/apko $apkover (+007-index-head-405-fallback)"
		fi
		export CGO_ENABLED=0
		go build -trimpath -o "$OUT/$tool" "$pkg"
	)
	echo "   built $OUT/$tool"
done
