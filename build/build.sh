#!/bin/bash
set -euo pipefail
# Cross-platform build script
GOOS=${1:-$(go env GOOS)}
GOARCH=${2:-$(go env GOARCH)}
BINARY=revora
BUILD_DIR=build

mkdir -p $BUILD_DIR
echo "Building for $GOOS/$GOARCH..."
GOOS=$GOOS GOARCH=$GOARCH go build -o $BUILD_DIR/${BINARY}-${GOOS}-${GOARCH} ./cmd/revora
echo "Done: $BUILD_DIR/${BINARY}-${GOOS}-${GOARCH}"