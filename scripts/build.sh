#!/bin/bash

# AITDD Build Script
# Usage: ./scripts/build.sh [version]

set -e

VERSION=${1:-"dev"}
BUILD_DIR="dist"
BINARY_NAME="aitdd"

echo "Building AITDD v${VERSION}..."

# Clean build directory
rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}

# Build backend
echo "Building backend..."
cd backend
GOOS=linux GOARCH=amd64 go build -ldflags="-X main.Version=${VERSION}" -o ../${BUILD_DIR}/${BINARY_NAME}-linux-amd64 ./cmd/aitdd
GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.Version=${VERSION}" -o ../${BUILD_DIR}/${BINARY_NAME}-darwin-amd64 ./cmd/aitdd
GOOS=darwin GOARCH=arm64 go build -ldflags="-X main.Version=${VERSION}" -o ../${BUILD_DIR}/${BINARY_NAME}-darwin-arm64 ./cmd/aitdd
GOOS=windows GOARCH=amd64 go build -ldflags="-X main.Version=${VERSION}" -o ../${BUILD_DIR}/${BINARY_NAME}-windows-amd64.exe ./cmd/aitdd
cd ..

# Build frontend
echo "Building frontend..."
cd frontend
npm ci
npm run build
mv dist ../${BUILD_DIR}/frontend
cd ..

# Copy additional files
echo "Copying additional files..."
cp README.md ${BUILD_DIR}/
cp -r docs ${BUILD_DIR}/
cp -r backend/migrations ${BUILD_DIR}/

# Create archives
echo "Creating archives..."
cd ${BUILD_DIR}
tar -czvf ${BINARY_NAME}-linux-amd64.tar.gz ${BINARY_NAME}-linux-amd64 frontend migrations docs README.md
tar -czvf ${BINARY_NAME}-darwin-amd64.tar.gz ${BINARY_NAME}-darwin-amd64 frontend migrations docs README.md
tar -czvf ${BINARY_NAME}-darwin-arm64.tar.gz ${BINARY_NAME}-darwin-arm64 frontend migrations docs README.md
zip -r ${BINARY_NAME}-windows-amd64.zip ${BINARY_NAME}-windows-amd64.exe frontend migrations docs README.md
cd ..

echo "Build complete!"
echo "Output files in ${BUILD_DIR}/"
ls -la ${BUILD_DIR}
