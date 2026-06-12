#!/usr/bin/env bash
set -e

# Usage: ./scripts/bump-version.sh [major|minor|patch]
# Bumps the VERSION file, commits, tags, and pushes.

PART="${1:-patch}"
VERSION_FILE="VERSION"

if [ ! -f "$VERSION_FILE" ]; then
  echo "VERSION file not found"
  exit 1
fi

CURRENT=$(cat "$VERSION_FILE" | tr -d '[:space:]')
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT"

case "$PART" in
  major)
    MAJOR=$((MAJOR + 1))
    MINOR=0
    PATCH=0
    ;;
  minor)
    MINOR=$((MINOR + 1))
    PATCH=0
    ;;
  patch)
    PATCH=$((PATCH + 1))
    ;;
  *)
    echo "Usage: $0 [major|minor|patch]"
    exit 1
    ;;
esac

NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"
echo "$NEW_VERSION" > "$VERSION_FILE"

git add "$VERSION_FILE"
git commit -m "chore(release): bump version to v${NEW_VERSION}"
git tag -a "v${NEW_VERSION}" -m "Release v${NEW_VERSION}"

echo "Bumped ${CURRENT} -> ${NEW_VERSION}"
echo "Run: git push origin main && git push origin v${NEW_VERSION}"
