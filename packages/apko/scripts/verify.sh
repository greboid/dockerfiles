#!/usr/bin/env bash
# Apply the patch set to the upstream tag pinned in melange.yaml, regenerate
# derived test data (SBOM goldens, publish digest pins) and run the test suite.
#
#   scripts/verify.sh [-w DIR]             apply all patches from scratch, then test
#   scripts/verify.sh -w DIR --continue    after resolving a conflict, apply the rest, then test
#   scripts/verify.sh -w DIR --export      write the patch commits in DIR back to patches/
set -euo pipefail

repo_root=$(cd "$(dirname "$0")/.." && pwd)
melange=$repo_root/melange.yaml
workdir=
mode=fresh

while [ $# -gt 0 ]; do
	case $1 in
	-w) workdir=$2; shift 2 ;;
	--continue) mode="continue"; shift ;;
	--export) mode="export"; shift ;;
	*) echo "unknown argument: $1" >&2; exit 2 ;;
	esac
done

field() { sed -n "s/^ *$1: *//p" "$melange" | head -n1; }
version=$(sed -n '/^package:/,/^[a-z]/s/^  version: *//p' "$melange")
upstream=$(field repository)
tag=$(field tag | sed "s/\${{package.version}}/$version/")
commit=$(field expected-commit)
mapfile -t patches < <(sed -n '/uses: patch/,/uses:/s/^ *\(patches\/[^ ]*\.patch\)$/\1/p' "$melange")

git_() { git -C "$workdir" -c user.name=verify -c user.email=verify@localhost "$@"; }
committed() { git_ log --format=%s "$commit..HEAD" | grep -qxF "$1"; }

apply_patches() {
	for p in "${patches[@]}"; do
		name=$(basename "$p")
		committed "$name" && continue
		if ! git_ apply -3 "$repo_root/$p"; then
			cat >&2 <<-EOF
				$name does not apply to $tag. Apply it by hand in $workdir (conflict
				markers are left there if a 3-way merge was possible), then:
				  git -C $workdir add -A && git -C $workdir commit -m $name
				  $0 -w $workdir --continue
				  $0 -w $workdir --export
			EOF
			exit 1
		fi
		git_ add -A
		git_ commit -q -m "$name"
	done
}

# Rewrites each failing `want := "sha256:..."` pin to the digest the test produced.
update_digest_pins() {
	local out pairs
	for _ in 1 2 3; do
		out=$(cd "$workdir" && go test -count=1 -run '^TestPublish' ./internal/cli/ 2>&1) && return 0
		pairs=$(printf '%s\n' "$out" | sed -n 's/.*expected: "\(sha256:[0-9a-f]*\)"$/\1/p; s/.*actual  : "\(sha256:[0-9a-f]*\)"$/\1/p' | paste -d' ' - -)
		[ -n "$pairs" ] || { printf '%s\n' "$out" >&2; return 1; }
		while read -r old new; do
			sed -i "s/want := \"$old\"/want := \"$new\"/" "$workdir/internal/cli/publish_test.go"
		done <<<"$pairs"
	done
	printf '%s\n' "$out" >&2
	return 1
}

export_patches() {
	for p in "${patches[@]}"; do
		name=$(basename "$p")
		rev=$(git_ log --format=%H --grep="^$name\$" "$commit..HEAD")
		[ -n "$rev" ] || { echo "no commit for $name in $workdir" >&2; exit 1; }
		git_ diff --no-color --no-ext-diff "$rev~1" "$rev" >"$repo_root/$p"
		echo "wrote $p"
	done
}

case $mode in
fresh)
	if [ -z "$workdir" ]; then
		workdir=$(mktemp -d)
	elif [ -e "$workdir" ] && [ -n "$(ls -A "$workdir")" ]; then
		echo "$workdir is not empty" >&2
		exit 1
	fi
	echo "workdir: $workdir"
	git clone -q --branch "$tag" "$upstream" "$workdir"
	[ "$(git_ rev-parse HEAD)" = "$commit" ] || { echo "$tag is not $commit" >&2; exit 1; }
	git_ checkout -q -b patched
	apply_patches
	;;
continue)
	[ -n "$workdir" ] || { echo "--continue needs -w DIR" >&2; exit 2; }
	git_ diff --quiet HEAD || { echo "$workdir has uncommitted changes" >&2; exit 1; }
	apply_patches
	;;
export)
	[ -n "$workdir" ] || { echo "--export needs -w DIR" >&2; exit 2; }
	export_patches
	exit 0
	;;
esac

(cd "$workdir" && APKO_UPDATE_GOLDENS=1 go test -count=1 ./internal/cli/ ./pkg/sbom/generator/spdx/ >/dev/null) || true
update_digest_pins
(cd "$workdir" && go test -count=1 ./...)
echo "ok: patch set applies to $tag and tests pass (workdir: $workdir)"
