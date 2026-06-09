#!/usr/bin/env bash
#
# Bumps the latest vX.Y.Z tag per Conventional Commits (feat -> minor,
# otherwise patch) and creates release via the GitLab API. A breaking change (`!`
# in a commit subject prefix) aborts so the major must be done manually. Only
# commit subjects are inspected: a `BREAKING CHANGE:` footer in a body is not
# detected. Driven by the weekly release schedule; see docs/release_process.md.
#
# GITLAB_TOKEN must be a real access token; CI_JOB_TOKEN tags do not trigger
# tag pipelines, so the release would never run.
#
set -euo pipefail

: "${GITLAB_TOKEN:?must be set (token glab uses to create the tag; not CI_JOB_TOKEN)}"
: "${CI_PROJECT_ID:?must be set (provided by GitLab CI)}"
: "${CI_COMMIT_SHA:?must be set (provided by GitLab CI)}"

# Only tag from a protected ref (main, or a future stable branch), so a release
# can't be cut from an arbitrary branch pipeline.
if [ "${CI_COMMIT_REF_PROTECTED:-}" != "true" ]; then
  echo "Refusing to tag: pipeline is not running on a protected ref." >&2
  exit 1
fi

git fetch --tags --quiet

# Highest stable vX.Y.Z (grep excludes pre-release tags; `|| true` lets an
# empty result fall through to the check below).
LATEST_TAG="$(git tag -l 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname \
  | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -n1 || true)"
if [ -z "$LATEST_TAG" ]; then
  echo "Could not find an existing vX.Y.Z tag to bump from." >&2
  exit 1
fi
echo "Latest release tag: $LATEST_TAG" >&2

# Releasable commits since the last tag, excluding merge commits
RELEASABLE="$(git log --no-merges --format='%h %s' "${LATEST_TAG}..HEAD")"

if [ -z "$RELEASABLE" ]; then
  echo "No releasable commits since ${LATEST_TAG}; skipping this week's release." >&2
  exit 0
fi

echo "Releasable commits since ${LATEST_TAG}:" >&2
echo "$RELEASABLE" >&2

# A `!` in the prefix marks a breaking change -> major, which must be done manually
if grep -qE '^[0-9a-f]+ [a-z]+(\([^)]*\))?!:' <<<"$RELEASABLE"; then
  echo "Breaking change detected since ${LATEST_TAG}; cut the major release manually." >&2
  exit 1
fi

# A `feat` bumps the minor version; anything else is a patch.
VERSION="${LATEST_TAG#v}"
MAJOR="${VERSION%%.*}"
REST="${VERSION#*.}"
MINOR="${REST%%.*}"
PATCH="${REST#*.}"

if grep -qE '^[0-9a-f]+ feat(\([^)]*\))?:' <<<"$RELEASABLE"; then
  NEW_TAG="v${MAJOR}.$((MINOR + 1)).0"
else
  NEW_TAG="v${MAJOR}.${MINOR}.$((PATCH + 1))"
fi
echo "New release tag: $NEW_TAG" >&2

# An existing tag means the desired state is already reached; exit 0 so a
# re-run does not fail.
if git rev-parse -q --verify "refs/tags/${NEW_TAG}" >/dev/null; then
  echo "Tag ${NEW_TAG} already exists; nothing to do." >&2
  exit 0
fi

# Create via the API so the token stays in the request auth, not a git remote URL.
glab api "projects/${CI_PROJECT_ID}/repository/tags" \
  -X POST \
  -f "tag_name=${NEW_TAG}" \
  -f "ref=${CI_COMMIT_SHA}" \
  -f "message=Release ${NEW_TAG}"

echo "Created ${NEW_TAG}. The tag pipeline will build and publish the release."
