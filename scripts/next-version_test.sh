#!/usr/bin/env bash
set -euo pipefail

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
version_script="$script_dir/next-version.sh"
tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT

new_repo() {
	local name=$1
	local repo="$tmp_dir/$name"

	git init -q "$repo"
	git -C "$repo" config user.name "release-test"
	git -C "$repo" config user.email "release-test@example.com"
	printf 'fixture\n' >"$repo/README.md"
	git -C "$repo" add README.md
	GIT_AUTHOR_DATE="2026-07-01T00:00:00Z" \
		GIT_COMMITTER_DATE="2026-07-01T00:00:00Z" \
		git -C "$repo" commit -q -m "fixture"
	printf '%s\n' "$repo"
}

tag_at() {
	local repo=$1
	local tag=$2
	local timestamp=$3

	GIT_COMMITTER_DATE="$timestamp" git -C "$repo" tag -a "$tag" -m "release $tag"
}

assert_version() {
	local expected=$1
	local repo=$2
	local major_release=${3:-false}
	local release_date=${4:-2026-07-17}
	local actual

	if ! actual=$(
		cd "$repo"
		RELEASE_DATE="$release_date" RELEASE_TZ=Asia/Shanghai \
			"$version_script" "$major_release"
	); then
		echo "next-version failed for $repo (major_release=$major_release)" >&2
		exit 1
	fi
	if [[ "$actual" != "$expected" ]]; then
		echo "expected $expected, got $actual (major_release=$major_release)" >&2
		exit 1
	fi
}

assert_failure() {
	local repo=$1
	local major_release=${2:-false}
	local release_date=${3:-2026-07-17}

	if (
		cd "$repo"
		RELEASE_DATE="$release_date" RELEASE_TZ=Asia/Shanghai \
			"$version_script" "$major_release"
	) >/dev/null 2>&1; then
		echo "expected next-version to fail for $repo (major_release=$major_release)" >&2
		exit 1
	fi
}

repo=$(new_repo initial)
assert_version "0.1.0" "$repo"
assert_version "1.0.0" "$repo" true

repo=$(new_repo same-day)
tag_at "$repo" v0.1.0 "2026-07-17T08:00:00+08:00"
assert_version "0.1.1" "$repo"
tag_at "$repo" v0.1.1 "2026-07-17T09:00:00+08:00"
assert_version "0.1.2" "$repo"

repo=$(new_repo next-day)
tag_at "$repo" v0.1.2 "2026-07-16T12:00:00+08:00"
assert_version "0.2.0" "$repo"

repo=$(new_repo timezone-boundary)
tag_at "$repo" v0.3.4 "2026-07-16T16:30:00Z"
assert_version "0.3.5" "$repo"

repo=$(new_repo first-major)
tag_at "$repo" v0.9.9 "2026-07-17T08:00:00+08:00"
assert_version "1.0.0" "$repo" true

repo=$(new_repo next-major)
tag_at "$repo" v1.4.3 "2026-07-17T08:00:00+08:00"
assert_failure "$repo" true

repo=$(new_repo invalid-input)
assert_failure "$repo" yes

echo "next-version tests passed"
