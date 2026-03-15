#!/bin/bash

# Build script for optimized LiteClaw Go binary

set -e

BINARY_NAME="liteclaw"
VERSION=${1:-"1.0.0"}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "=== LiteClaw Go Optimization Build Script ==="
echo "Version: $VERSION"
echo "Commit: $GIT_COMMIT"
echo ""

# Clean previous builds
echo "Cleaning previous builds..."
rm -rf bin/
mkdir -p bin/

# Build flags
LDFLAGS="-s -w -X main.version=$VERSION -X main.buildDate=$BUILD_TIME -X main.commit=$GIT_COMMIT -extldflags '-static'"
GCFLAGS="-l=4 -B"
BUILDTAGS="netgo,osusergo,static_build"

echo ""
echo "=== Building Optimized Binary ==="

# Step 1: Build with all optimizations
echo "1. Building with optimizations..."
CGO_ENABLED=0 go build \
    -ldflags="$LDFLAGS" \
    -gcflags="$GCFLAGS" \
    -asmflags="-trimpath=$(pwd)" \
    -trimpath \
    -tags="$BUILDTAGS" \
    -o bin/$BINARY_NAME \
    .

SIZE=$(ls -lh bin/$BINARY_NAME | awk '{print $5}')
echo "   Binary size: $SIZE"

# Step 2: Apply UPX compression (if available)
echo ""
echo "2. Applying UPX compression..."
if command -v upx &> /dev/null; then
    cp bin/$BINARY_NAME bin/${BINARY_NAME}-uncompressed
    upx --best --lzma bin/$BINARY_NAME -o bin/${BINARY_NAME}-compressed 2>/dev/null
    
    # Use compressed version as main binary
    mv bin/${BINARY_NAME}-compressed bin/${BINARY_NAME}
    
    SIZE_COMPRESSED=$(ls -lh bin/$BINARY_NAME | awk '{print $5}')
    SIZE_UNCOMPRESSED=$(ls -lh bin/${BINARY_NAME}-uncompressed | awk '{print $5}')
    
    echo "   Uncompressed: $SIZE_UNCOMPRESSED"
    echo "   Compressed:   $SIZE_COMPRESSED"
    
    # Keep uncompressed for comparison
    mv bin/${BINARY_NAME}-uncompressed bin/${BINARY_NAME}-debug
else
    echo "   UPX not found, skipping compression"
    echo "   Install with: apt-get install upx-ucl or brew install upx"
fi

# Step 3: Build for multiple platforms
echo ""
echo "3. Building for multiple platforms..."

# Linux AMD64
echo "   Building for Linux AMD64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="$LDFLAGS" \
    -gcflags="$GCFLAGS" \
    -trimpath \
    -tags="$BUILDTAGS" \
    -o bin/${BINARY_NAME}-linux-amd64 \
    .

# Linux ARM64
echo "   Building for Linux ARM64..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
    -ldflags="$LDFLAGS" \
    -gcflags="$GCFLAGS" \
    -trimpath \
    -tags="$BUILDTAGS" \
    -o bin/${BINARY_NAME}-linux-arm64 \
    .

# macOS AMD64
echo "   Building for macOS AMD64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
    -ldflags="$LDFLAGS" \
    -gcflags="$GCFLAGS" \
    -trimpath \
    -o bin/${BINARY_NAME}-darwin-amd64 \
    .

# macOS ARM64 (M1/M2)
echo "   Building for macOS ARM64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
    -ldflags="$LDFLAGS" \
    -gcflags="$GCFLAGS" \
    -trimpath \
    -o bin/${BINARY_NAME}-darwin-arm64 \
    .

# Windows AMD64
echo "   Building for Windows AMD64..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
    -ldflags="$LDFLAGS" \
    -gcflags="$GCFLAGS" \
    -trimpath \
    -o bin/${BINARY_NAME}-windows-amd64.exe \
    .

# Step 4: Apply UPX to all platforms
if command -v upx &> /dev/null; then
    echo ""
    echo "4. Compressing all platform binaries..."
    for file in bin/${BINARY_NAME}-*; do
        if [[ "$file" != *".exe" ]]; then
            upx --best --lzma "$file" 2>/dev/null || true
        fi
    done
fi

# Step 5: Display results
echo ""
echo "=== Build Results ==="
echo ""
ls -lh bin/

echo ""
echo "=== Binary Size Comparison ==="
echo ""
MAIN_SIZE=$(stat -c%s bin/$BINARY_NAME 2>/dev/null || stat -f%z bin/$BINARY_NAME)
echo "Main binary:    $(ls -lh bin/$BINARY_NAME | awk '{print $5}')"
echo "Target (Rust):  2.8M"
echo "Difference:     $(echo "scale=1; $MAIN_SIZE / 1024 / 1024" | bc)M"

# Step 6: Memory optimization check
echo ""
echo "=== Memory Optimization Tips ==="
echo ""
echo "To achieve < 5MB memory usage:"
echo "1. Set GC percent: debug.SetGCPercent(50)"
echo "2. Set memory limit: debug.SetMemoryLimit(5 << 20)"
echo "3. Use sync.Pool for reusable objects"
echo "4. Pre-allocate slices and maps"
echo "5. Avoid unnecessary allocations"
echo ""
echo "Run with:"
echo "  GODEBUG=gctrace=1 ./bin/$BINARY_NAME"
echo ""

# Step 7: Final verification
echo "=== Verification ==="
echo ""
echo "Binary size target: < 3 MB (with UPX)"
echo "Memory target: < 5 MB"
echo ""
echo "✓ Build complete!"
echo ""
echo "To test:"
echo "  ./bin/$BINARY_NAME --version"
echo "  ./bin/$BINARY_NAME server --port 8080"
echo ""
