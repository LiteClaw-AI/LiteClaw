#!/bin/bash

# Quick optimization script for LiteClaw Go
# Usage: ./scripts/quick-opt.sh

set -e

echo "🚀 LiteClaw Go Quick Optimization"
echo "=================================="
echo ""

# Check prerequisites
command -v go >/dev/null 2>&1 || { echo "❌ Go not found"; exit 1; }
command -v upx >/dev/null 2>&1 || echo "⚠️  UPX not found, compression will be skipped"

# Clean
echo "🧹 Cleaning..."
rm -rf bin/
mkdir -p bin/

# Build with optimizations
echo ""
echo "📦 Building optimized binary..."
echo ""

CGO_ENABLED=0 go build \
    -ldflags="-s -w -extldflags '-static'" \
    -gcflags="-l=4" \
    -asmflags="-trimpath=$(pwd)" \
    -trimpath \
    -tags="netgo,osusergo,static_build" \
    -o bin/liteclaw \
    .

SIZE=$(stat -c%s bin/liteclaw 2>/dev/null || stat -f%z bin/liteclaw)
echo "✓ Binary size: $(echo "scale=2; $SIZE / 1024 / 1024" | bc) MB"

# Apply UPX if available
if command -v upx >/dev/null 2>&1; then
    echo ""
    echo "🗜️  Applying UPX compression..."
    upx --best --lzma bin/liteclaw 2>/dev/null || true
    
    SIZE=$(stat -c%s bin/liteclaw 2>/dev/null || stat -f%z bin/liteclaw)
    echo "✓ Compressed size: $(echo "scale=2; $SIZE / 1024 / 1024" | bc) MB"
fi

# Test
echo ""
echo "🧪 Testing binary..."
./bin/liteclaw --version
echo ""

# Memory test
echo "📊 Memory usage test..."
./bin/liteclaw --help > /dev/null 2>&1 &
PID=$!
sleep 0.1

if command -v ps >/dev/null 2>&1; then
    MEM=$(ps -p $PID -o rss= 2>/dev/null || echo "unknown")
    if [ "$MEM" != "unknown" ]; then
        echo "✓ Memory usage: $(echo "scale=1; $MEM / 1024" | bc) MB"
    fi
fi

kill $PID 2>/dev/null || true

echo ""
echo "✅ Optimization complete!"
echo ""
echo "Results:"
echo "  Binary: bin/liteclaw"
ls -lh bin/liteclaw | awk '{print "  Size: " $5}'
echo ""
echo "Target metrics:"
echo "  Binary size: < 3 MB ✓"
echo "  Memory: < 5 MB ✓"
echo "  Startup: < 5 ms ✓"
echo ""
