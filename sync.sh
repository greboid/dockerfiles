#!/bin/bash
#
# sync.sh - rebuild the read-only monorepo view of the Forgejo organisations.
#
# Enumerates every repository in the configured organisations, shallow-fetches
# each default branch and snapshots it into this repository under <org>/<name>/.
# A single commit is pushed only when the resulting tree differs from the
# current one, so this repository is generated output: never commit to it
# directly, and never push to it from anywhere but this script.
#
# Requirements: git, curl, jq, bash, awk.
#
# Environment:
#   FORGEJO_URL    instance base URL          (default https://git.mouse-lake.ts.net)
#   SYNC_ORGS      organisations to sync      (default "containers packages")
#   SYNC_BRANCH    branch the view is pushed  (default master)
#   PUSH_REPO      target slug owner/repo     (default: parsed from origin)
#   FORGEJO_TOKEN  API token with repo write; falls back to the PAT file
#                  in the repository root.

set -euo pipefail

FORGEJO_URL="${FORGEJO_URL:-https://git.mouse-lake.ts.net}"
SYNC_ORGS="${SYNC_ORGS:-containers packages}"
SYNC_BRANCH="${SYNC_BRANCH:-master}"

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

if [ -z "${FORGEJO_TOKEN:-}" ] && [ -f "$REPO_ROOT/PAT" ]; then
	FORGEJO_TOKEN="$(tr -d '[:space:]' <"$REPO_ROOT/PAT")"
fi
if [ -z "${FORGEJO_TOKEN:-}" ]; then
	echo "error: FORGEJO_TOKEN is not set and no PAT file was found in $REPO_ROOT" >&2
	exit 2
fi
export GIT_TERMINAL_PROMPT=0

if [ -z "${PUSH_REPO:-}" ]; then
	url="$(git config --get remote.origin.url || true)"
	if [ -z "$url" ]; then
		echo "error: no origin remote and PUSH_REPO is not set" >&2
		exit 2
	fi
	url="${url%.git}"
	url="${url#*://}" # ssh://git@host/owner/repo -> git@host/owner/repo
	url="${url#*@}"   # drop user@ / user:token@
	url="${url/:/\/}" # scp-style git@host:owner/repo -> host/owner/repo
	PUSH_REPO="$(basename "${url%/*}")/$(basename "$url")"
fi

API="$FORGEJO_URL/api/v1"
AUTH_USER="$(curl -fsS -H "authorization: token $FORGEJO_TOKEN" "$API/user" | jq -r .login)"
AUTHED_BASE="https://${AUTH_USER}:${FORGEJO_TOKEN}@${FORGEJO_URL#https://}"

tmp="$(mktemp -d)"
entries="$tmp/entries.tsv"   # full_name, name, branch, clone_url, empty, archived
shas="$tmp/shas.tsv"         # org, name, source commit (per build attempt)
trap 'rm -rf "$tmp"' EXIT

echo "Enumerating repositories in: $SYNC_ORGS"
: >"$entries"
for org in $SYNC_ORGS; do
	page=1
	while :; do
		resp="$(curl -fsS -H "authorization: token $FORGEJO_TOKEN" "$API/orgs/$org/repos?limit=50&page=$page")"
		count="$(jq 'length' <<<"$resp")"
		[ "$count" -eq 0 ] && break
		jq -r '.[] | [.full_name, .name, .default_branch, .clone_url, (.empty|tostring), (.archived|tostring)] | @tsv' \
			<<<"$resp" >>"$entries"
		page=$((page + 1))
	done
done

# Drop empty repositories; keep everything else (archived included).
skipped="$(awk -F'\t' '$5 == "true" {print "  " $1}' "$entries")"
awk -F'\t' '$5 != "true"' "$entries" >"$entries.sync"
total="$(wc -l <"$entries.sync")"
summary="$(cut -f1 "$entries.sync" | cut -d/ -f1 | sort | uniq -c | awk '{printf "%s(%d) ", $2, $1}')"
summary="${summary% }"
echo "Found: $summary"

base() {
	# Tip of the view branch: local ref if present, otherwise the checkout HEAD.
	git rev-parse -q --verify "refs/heads/$SYNC_BRANCH^{commit}" 2>/dev/null || git rev-parse HEAD
}

build_tree() {
	# $1 = base commit; prints the snapshot tree built on top of it.
	local idx base org name branch url sha tree
	idx="$tmp/index"
	base="$1"
	: >"$shas"

	git read-tree --index-output="$idx" "$base"
	# Replace the previously synced prefixes with a fresh snapshot.
	GIT_INDEX_FILE="$idx" git rm -r -q --cached $SYNC_ORGS 2>/dev/null || true

	while IFS=$'\t' read -r _full name branch clone_url _empty _archived; do
		org="${_full%%/*}"
		url="${clone_url/https:\/\//https:\/\/${AUTH_USER}:${FORGEJO_TOKEN}@}"
		git fetch -q --no-tags --depth 1 "$url" "$branch"
		sha="$(git rev-parse -q --verify FETCH_HEAD)"
		tree="$(git rev-parse -q --verify "$sha^{tree}")"
		GIT_INDEX_FILE="$idx" git read-tree --prefix="$org/$name/" "$tree"
		printf '%s\t%s\t%s\n' "$org" "$name" "${sha:0:12}" >>"$shas"
	done <"$entries.sync"

	GIT_INDEX_FILE="$idx" git write-tree
}

commit_message() {
	printf 'Sync: %s\n\n' "$summary"
	printf 'Generated snapshot of the %s organisations.\n' "$SYNC_ORGS"
	printf 'This repository is view only; all changes belong in the organisation repositories.\n\n'
	sort "$shas" | while IFS=$'\t' read -r org name sha; do
		printf '%-32s %s\n' "$org/$name" "$sha"
	done
}

commit_and_push() {
	local attempts=0 attempt_base tree commit
	attempt_base="$1"
	while :; do
		attempts=$((attempts + 1))
		tree="$(build_tree "$attempt_base")"

		if [ "$tree" = "$(git rev-parse -q --verify "$attempt_base^{tree}")" ]; then
			echo "View is up to date ($total repositories); nothing to commit."
			echo "$attempt_base"
			return 0
		fi

		commit="$(commit_message | git commit-tree "$tree" -p "$attempt_base")"
		echo "Pushing snapshot $commit"
		if git push -q "$AUTHED_BASE/${PUSH_REPO}.git" "$commit:refs/heads/$SYNC_BRANCH"; then
			echo "$commit"
			return 0
		fi

		if [ "$attempts" -ge 3 ]; then
			echo "error: giving up after $attempts attempts (push kept being rejected)" >&2
			exit 1
		fi
		echo "Push rejected (branch moved); rebuilding on top of the new tip"
		git fetch -q --no-tags --depth 1 "$AUTHED_BASE/${PUSH_REPO}.git" "$SYNC_BRANCH"
		attempt_base="$(git rev-parse -q --verify FETCH_HEAD)"
	done
}

commit="$(commit_and_push "$(base)")"

if [ -z "${CI:-}" ]; then
	git update-ref "refs/heads/$SYNC_BRANCH" "$commit"
	git reset -q --hard
	echo "Local worktree updated to $commit"
fi

if [ -n "$skipped" ]; then
	echo "Skipped empty repositories:"
	echo "$skipped"
fi
echo "Done: $summary"
