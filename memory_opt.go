package main

import (
	"runtime/debug"
	"time"
)

func init() {
	// Optimize GC for low memory usage
	// Lower value = more frequent GC = less memory
	debug.SetGCPercent(50) // Default is 100

	// Limit memory usage (Go 1.19+)
	// Set a soft memory limit of 5MB
	debug.SetMemoryLimit(5 * 1024 * 1024)

	// Optimize for single-threaded performance
	runtime.GOMAXPROCS(1) // Uncomment for single-core systems
}

// Compile-time optimizations
// Build with: go build -ldflags="-s -w" -gcflags="-l=4 -B"

/*
Memory Optimization Strategies:

1. GC Tuning (debug.SetGCPercent)
   - Default: 100 (balance between CPU and memory)
   - Lower (e.g., 50): More frequent GC, less memory
   - Higher (e.g., 200): Less frequent GC, more memory

2. Memory Limit (debug.SetMemoryLimit) - Go 1.19+
   - Soft limit on heap size
   - Triggers GC when approaching limit
   - Helps prevent OOM in containerized environments

3. Reduce Allocations
   - Use sync.Pool for reusable objects
   - Pre-allocate slices and maps
   - Avoid interface{} when possible

4. String Optimization
   - Use strings.Builder for concatenation
   - Intern repeated strings
   - Use []byte when possible

5. Struct Optimization
   - Order fields by size (largest first)
   - Use smaller types (int32 vs int64)
   - Avoid unnecessary pointers

6. Slice Optimization
   - Pre-allocate with make([]T, 0, cap)
   - Reuse slices with [:0]
   - Use copy instead of append when possible

7. Map Optimization
   - Pre-allocate with make(map[K]V, size)
   - Use struct keys when possible
   - Consider arrays for small, fixed-size maps

8. Interface Optimization
   - Use concrete types when possible
   - Small interfaces (1-3 methods)
   - Avoid empty interface (interface{})

Binary Size Optimization:

1. Build Flags
   -ldflags "-s -w"        # Strip debug info
   -gcflags "-l=4 -B"      # Inlining optimization
   -asmflags "-trimpath"   # Remove paths
   -trimpath               # Remove all paths

2. Build Tags
   -tags "netgo,osusergo,static_build"
   CGO_ENABLED=0

3. UPX Compression
   upx --best --lzma binary

4. Remove Dependencies
   go mod tidy
   Check with: go list -m all

Target Metrics:
- Binary size: < 3 MB (with UPX)
- Memory usage: < 5 MB
- Startup time: < 5 ms
*/
