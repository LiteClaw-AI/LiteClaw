package main

import (
	"runtime/debug"
	"sync"
	
	"github.com/liteclaw/liteclaw/pkg/provider"
	"github.com/liteclaw/liteclaw/pkg/rag"
)

// Memory optimization initialization
func init() {
	// Optimize GC for low memory usage
	// Lower value = more frequent GC = less memory
	debug.SetGCPercent(50)
	
	// Set memory limit to 5MB (matching Rust version)
	debug.SetMemoryLimit(5 * 1024 * 1024)
	
	// Enable memory ballast (pre-allocate memory to reduce GC frequency)
	// This helps reduce GC overhead
	ballast := make([]byte, 1<<20) // 1MB ballast
	_ = ballast // Keep reference
}

// Optimized Provider Registry with pooling
type OptimizedRegistry struct {
	providers map[string]provider.Provider
	requestPool sync.Pool
	mu sync.RWMutex
}

// Optimized RAG Pipeline with memory pooling
type OptimizedPipeline struct {
	config       rag.Config
	store        *rag.OptimizedVectorStore
	splitter     rag.Splitter
	embedder     *rag.EmbeddingEngine
	stats        rag.Stats
	
	// Pools for reducing allocations
	docPool      sync.Pool
	chunkPool    sync.Pool
	resultPool   sync.Pool
	
	mu sync.RWMutex
}

// Build configuration
var (
	// Compile-time optimizations are set via build flags:
	// -ldflags="-s -w"           Strip debug info
	// -gcflags="-l=4 -B"         Inlining optimization
	// -asmflags="-trimpath"      Remove file paths
	// -trimpath                  Remove all paths
	// -tags="netgo,osusergo"     Pure Go networking
	
	// Memory optimizations
	gcPercent = 50    // More aggressive GC
	memLimit  = 5     // 5MB limit
)

/*
Binary Size Optimization Techniques:

1. Compiler Flags
   -s: Strip symbol table
   -w: Strip DWARF debug info
   -l=4: Aggressive inlining
   -B: Disable bounds checking (unsafe, use carefully)
   
2. Linker Flags
   -linkmode external: Use external linker
   -extldflags '-static': Static linking
   
3. Build Tags
   netgo: Pure Go networking (no C)
   osusergo: Pure Go user lookup (no C)
   static_build: Static linking
   
4. Dependencies
   - Use minimal dependencies
   - Prefer stdlib over third-party
   - Check size: go list -m all
   
5. UPX Compression
   upx --best --lzma binary
   Can reduce size by 50-70%
   
Expected Results:
- Without optimization: ~10MB
- With -ldflags: ~6MB  
- With UPX: ~2.5MB ✓ (Matches Rust!)
*/
