#!/usr/bin/env bash
# Print the next release version (without the leading "v").
#
# Default progression uses RELEASE_TZ (Asia/Shanghai by default):
#   - first release: 0.1.0
#   - same natural day: increment patch
#   - later natural day: increment minor and reset patch
#
# Pass "true" to cut v1.0.0 from a v0 release. Go modules require a /vN
# module-path suffix for v2+, so those releases need a dedicated migration.
set -euo pipefail

major_release=${1:-false}
case "$major_release" in
	true | false) ;;
	*)
		echo "error: major_release must be true or false, got '$major_release'" >&2
		exit 2
		;;
esac

export TZ=${RELEASE_TZ:-Asia/Shanghai}
today=${RELEASE_DATE:-$(date +%Y-%m-%d)}
if ! [[ "$today" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]]; then
	echo "error: RELEASE_DATE must be YYYY-MM-DD, got '$today'" >&2
	exit 2
fi

latest=$(git tag --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | head -1)
if [[ -z "$latest" ]]; then
	if [[ "$major_release" == true ]]; then
		echo "1.0.0"
	else
		echo "0.1.0"
	fi
	exit 0
fi

if ! [[ "$latest" =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
	echo "error: latest release tag is not vMAJOR.MINOR.PATCH: $latest" >&2
	exit 2
fi
major=${BASH_REMATCH[1]}
minor=${BASH_REMATCH[2]}
patch=${BASH_REMATCH[3]}

if [[ "$major_release" == true ]]; then
	if ((major >= 1)); then
		next_major=$((major + 1))
		echo "error: v${next_major}.0.0 requires a /v${next_major} module-path migration before release" >&2
		exit 2
	fi
	echo "1.0.0"
	exit 0
fi

last_day=$(git for-each-ref \
	--format='%(creatordate:format-local:%Y-%m-%d)' \
	"refs/tags/$latest")
if [[ -z "$last_day" ]]; then
	echo "error: could not read creation date for $latest" >&2
	exit 2
fi

if [[ "$last_day" == "$today" ]]; then
	echo "${major}.${minor}.$((patch + 1))"
else
	echo "${major}.$((minor + 1)).0"
fi
