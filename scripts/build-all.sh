#!/bin/bash
set -e

VERSION=${1:-dev}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION=$(go version | cut -d' ' -f3)
CMD_DIR=./cmd/gonzofk
OUTPUT_DIR="dist"

LDFLAGS="-s -w \
-X main.version=${VERSION} \
-X main.commit=${COMMIT} \
-X main.buildTime=${BUILD_TIME} \
-X main.goVersion=${GO_VERSION}"

echo "Building gonzofk ${VERSION} (${COMMIT})..."

rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

build() {
    local goos=$1 goarch=$2 ext=$3
    echo "  -> ${goos}/${goarch}"
    CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build -trimpath \
        -ldflags "$LDFLAGS" \
        -o "$OUTPUT_DIR/gonzofk-${goos}-${goarch}${ext}" "$CMD_DIR"
}

build linux  amd64 ""
build linux  arm64 ""
build darwin amd64 ""
build darwin arm64 ""
build windows amd64 ".exe"

echo ""
echo "Build complete:"
ls -lh "$OUTPUT_DIR/"
