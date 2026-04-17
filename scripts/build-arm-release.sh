#!/usr/bin/env bash
# Build the ARM binary for a specific release version.
# Usage: ./scripts/build-arm-release.sh v0.2.4

set -e

if [ $# -ne 1 ]; then
  echo "Usage: $0 <version>"
  echo "Example: $0 v0.2.4"
  exit 1
fi

VERSION="$1"
BUILD_DIR="build"
OUTPUT="${BUILD_DIR}/luckyclaw-linux-arm"

mkdir -p "${BUILD_DIR}"

GIT_COMMIT=$(git rev-parse --short=8 HEAD 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u +'%Y-%m-%dT%H:%M:%S%z' 2>/dev/null || echo "unknown")
GO_VERSION=$(go version | awk '{print $3}')

echo "Building luckyclaw for ARM (release)…"
echo "Version:        ${VERSION}"
echo "Commit:         ${GIT_COMMIT}"
echo "Build time:     ${BUILD_TIME}"
echo "Go version:     ${GO_VERSION}"

CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 \
  go build -ldflags \
    "-s -w \
    -X main.version=${VERSION} \
    -X main.gitCommit=${GIT_COMMIT} \
    -X main.buildTime=${BUILD_TIME} \
    -X main.goVersion=${GO_VERSION}" \
  -o "${OUTPUT}" ./cmd/luckyclaw

FILE_OUTPUT=$(file "${OUTPUT}")
if ! echo "${FILE_OUTPUT}" | grep -q 'ARM'; then
  echo "ERROR: Built binary is not ARM:"
  echo "  ${FILE_OUTPUT}"
  exit 1
fi

echo ""
echo "Build complete: ${OUTPUT}"
echo ""
echo "Next steps:"
echo "  ./scripts/sync-overlay.sh"
echo "  cd luckfox-pico-sdk && ./build.sh"
