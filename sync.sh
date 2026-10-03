#!/bin/bash
#
# sync.sh - rebuild the read-only monorepo view of the Forgejo organisations.
#
# Enumerates every repository in the configured organisations, shallow-fetches
# each default branch, materialises them into the worktree under
# <org>/<name>/ and commits the result as a single squashed snapshot.
# A commit is pushed only when the org content changed, so this repository is
# generated output: never edit containers/ or packages/ directly, and never
# push to this repository from anywhere but this script.
#
# The org directories are wiped and rebuilt from the fetched commits on every
# run; infra files (README, sync.sh, workflows, ...) are never touched by the
# sync and are maintained by ordinary commits. Before committing, the run
# verifies the infra files are still present and aborts otherwise.
#
# Requirements: git, curl, jq, bash, awk, tar.
#
# Environment:
#   FORGEJO_URL     instance base URL          (default https://git.mouse-lake.ts.net)
#   SYNC_ORGS       organisations to sync      (default "containers packages")
#   SYNC_BRANCH     branch the view is pushed  (default master)
#   PUSH_REPO       target slug owner/repo     (default: parsed from origin)
#   REQUIRED_INFRA  files that must exist before a commit is made
#                   (default "README.md sync.sh .forgejo/workflows/sync.yml")
#   FORGEJO_TOKEN   API token with repo write; falls back to the PAT file
#                   in the repository root.

set -euo pipefail

FORGEJO_URL="${FORGEJO_URL:-https://git.mouse-lake.ts.net}"
SYNC_ORGS="${SYNC_ORGS:-containers packages}"
SYNC_BRANCH="${SYNC_BRANCH:-master}"
REQUIRED_INFRA="${REQUIRED_INFRA:-README.md sync.sh .forgejo/workflows/sync.yml}"

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
user_json="$(curl -fsS -H "authorization: token $FORGEJO_TOKEN" "$API/user")"
AUTH_USER="$(jq -r .login <<<"$user_json")"
AUTH_EMAIL="$(jq -r .email <<<"$user_json")"
AUTHED_BASE="https://${AUTH_USER}:${FORGEJO_TOKEN}@${FORGEJO_URL#https://}"

# Commit identity for CI environments with no configured user.
if [ -z "$(git config user.name || true)" ]; then
	export GIT_AUTHOR_NAME="$AUTH_USER" GIT_COMMITTER_NAME="$AUTH_USER"
fi
if [ -z "$(git config user.email || true)" ]; then
	export GIT_AUTHOR_EMAIL="$AUTH_EMAIL" GIT_COMMITTER_EMAIL="$AUTH_EMAIL"
fi

tmp="$(mktemp -d)"
entries="$tmp/entries.tsv" # full_name, name, branch, clone_url, empty, archived
shas="$tmp/shas.tsv"       # org, name, fetched commit
trap 'rm -rf "$tmp"' EXIT

echo "Enumerating repositories in: $SYNC_ORGS" >&2
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
echo "Found: $summary" >&2

fetch_repos() {
	# Shallow-fetch every repository and record its tip commit. The objects
	# stay in the local store, so a later retry can re-extract without
	# re-fetching.
	: >"$shas"
	local _full name branch clone_url org url sha
	while IFS=$'\t' read -r _full name branch clone_url _empty _archived; do
		org="${_full%%/*}"
		url="${clone_url/https:\/\//https:\/\/${AUTH_USER}:${FORGEJO_TOKEN}@}"
		git fetch -q --no-tags --depth 1 "$url" "$branch"
		sha="$(git rev-parse -q --verify FETCH_HEAD)"
		printf '%s\t%s\t%s\n' "$org" "$name" "$sha" >>"$shas"
	done <"$entries.sync"
}

extract_repos() {
	# Wipe the org directories and materialise each fetched commit into the
	# worktree. Everything is rebuilt from scratch, so repositories deleted
	# or emptied upstream disappear here too.
	local org name sha
	for org in $SYNC_ORGS; do
		rm -rf "$org"
	done
	while IFS=$'\t' read -r org name sha; do
		mkdir -p "$org/$name"
		git archive --format=tar "$sha" | tar -x -C "$org/$name"
	done <"$shas"
}

verify_infra() {
	# The infra files must exist in the worktree before anything is committed:
	# they are outside the sync's write path, so them going missing means
	# something else is badly wrong.
	local bad=0 f
	for f in $REQUIRED_INFRA; do
		if [ ! -e "$f" ]; then
			echo "error: infra file '$f' is missing from the worktree" >&2
			bad=1
		fi
	done
	if [ "$bad" -ne 0 ]; then
		echo "error: refusing to commit a snapshot without infra files" >&2
		exit 1
	fi
}

commit_message() {
	printf 'Sync: %s\n\n' "$summary"
	printf 'Generated snapshot of the %s organisations.\n' "$SYNC_ORGS"
	printf 'This repository is view only; all changes belong in the organisation repositories.\n\n'
	sort "$shas" | while IFS=$'\t' read -r org name sha; do
		printf '%-32s %s\n' "$org/$name" "${sha:0:12}"
	done
}

sync_view() {
	local attempts=0 tip
	while :; do
		attempts=$((attempts + 1))

		# Always build on the current server-side tip of the view branch.
		tip="$(git ls-remote -q "$AUTHED_BASE/${PUSH_REPO}.git" "refs/heads/$SYNC_BRANCH" | cut -f1)"
		if [ -n "$tip" ] && [ "$tip" != "$(git rev-parse HEAD)" ]; then
			git fetch -q --no-tags --depth 1 "$AUTHED_BASE/${PUSH_REPO}.git" "$SYNC_BRANCH"
			git reset -q --hard FETCH_HEAD
		fi
		verify_infra

		extract_repos
		# -f: files a synced repository ignores itself must still land in the view.
		git add -A -f -- $SYNC_ORGS

		if git diff --cached --quiet; then
			echo "View is up to date ($total repositories); nothing to commit." >&2
			return 0
		fi

		git commit -q -m "$(commit_message)"
		echo "Pushing snapshot $(git rev-parse HEAD)" >&2
		if git push -q "$AUTHED_BASE/${PUSH_REPO}.git" "HEAD:refs/heads/$SYNC_BRANCH"; then
			return 0
		fi

		if [ "$attempts" -ge 3 ]; then
			echo "error: giving up after $attempts attempts (push kept being rejected)" >&2
			exit 1
		fi
		# Retry re-extracts from the already-fetched objects; no re-fetch needed.
		echo "Push rejected (branch moved); rebuilding on top of the new tip" >&2
	done
}

fetch_repos
sync_view

if [ -n "$skipped" ]; then
	echo "Skipped empty repositories:"
	echo "$skipped"
fi
echo "Done: $summary"
