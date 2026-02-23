#!/bin/bash

# AITDD Release Script
# Usage: ./scripts/release.sh <version>

set -e

if [ -z "$1" ]; then
    echo "Usage: ./scripts/release.sh <version>"
    echo "Example: ./scripts/release.sh v1.0.0"
    exit 1
fi

VERSION=$1
echo "Preparing release ${VERSION}..."

# Validate version format
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Version must be in format v1.0.0"
    exit 1
fi

# Check for uncommitted changes
if ! git diff-index --quiet HEAD --; then
    echo "Error: You have uncommitted changes"
    exit 1
fi

# Run tests
echo "Running tests..."
cd backend && go test ./... && cd ..
cd frontend && npm test && cd .. 2>/dev/null || true

# Build
echo "Building..."
./scripts/build.sh ${VERSION#v}

# Create git tag
echo "Creating git tag..."
git tag -a ${VERSION} -m "Release ${VERSION}"
git push origin ${VERSION}

# Create GitHub release
echo "Creating GitHub release..."
gh release create ${VERSION} \
    --title "AITDD ${VERSION}" \
    --notes "## What's Changed

### New Features
-

### Bug Fixes
-

### Improvements
-

**Full Changelog**: https://github.com/lujunchen71/AITDD/compare/\$(git describe --tags --abbrev=0 ${VERSION}^)..${VERSION}" \
    dist/aitdd-*.tar.gz \
    dist/aitdd-*.zip

echo "Release ${VERSION} complete!"
echo "Download from: https://github.com/lujunchen71/AITDD/releases/tag/${VERSION}"
