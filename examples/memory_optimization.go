package main

import (
	"fmt"
	"os"
	"runtime"
	"time"
	"unsafe"
)

// Memory optimization examples

func main() {
	fmt.Println("=== LiteClaw Go Memory Optimization Demo ===\n")

	// Show initial memory stats
	printMemStats("Initial")

	// 1. String interning
	demoStringInterning()

	// 2. Sync.Pool for object reuse
	demoSyncPool()

	// 3. Pre-allocated slices
	demoPreallocation()

	// 4. Value types vs pointers
	demoValueTypes()

	// 5. Memory pooling
	demoMemoryPooling()

	// Final stats
	printMemStats("Final")

	// Optimization tips
	printOptimizationTips()
}

// String interning to reduce memory
var stringPool = make(map[string]string)

func internString(s string) string {
	if interned, exists := stringPool[s]; exists {
		return interned
	}
	stringPool[s] = s
	return s
}

func demoStringInterning() {
	fmt.Println("1. String Interning")
	fmt.Println("   Reduces memory for repeated strings")

	// Without interning
	strings1 := make([]string, 1000)
	for i := range strings1 {
		strings1[i] = "repeated_string_value"
	}

	// With interning
	strings2 := make([]string, 1000)
	for i := range strings2 {
		strings2[i] = internString("repeated_string_value")
	}

	fmt.Printf("   Saved: ~%d bytes\n\n", 
		(len("repeated_string_value") * 1000) - 
		(len("repeated_string_value") + len(stringPool)))
}

// Sync.Pool for object reuse
import "sync"

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 1024)
	},
}

func demoSyncPool() {
	fmt.Println("2. Sync.Pool for Object Reuse")

	// Get from pool
	buf := bufferPool.Get().([]byte)
	
	// Use buffer
	for i := range buf[:10] {
		buf[i] = byte(i)
	}

	// Return to pool
	bufferPool.Put(buf)

	fmt.Println("   Reuses objects, reduces GC pressure")
	fmt.Printf("   Pool size: flexible (managed by GC)\n\n")
}

// Pre-allocation
func demoPreallocation() {
	fmt.Println("3. Pre-allocated Slices")

	// Without pre-allocation
	start1 := time.Now()
	var slice1 []int
	for i := 0; i < 10000; i++ {
		slice1 = append(slice1, i)
	}
	dur1 := time.Since(start1)

	// With pre-allocation
	start2 := time.Now()
	slice2 := make([]int, 0, 10000)
	for i := 0; i < 10000; i++ {
		slice2 = append(slice2, i)
	}
	dur2 := time.Since(start2)

	fmt.Printf("   Without pre-alloc: %v\n", dur1)
	fmt.Printf("   With pre-alloc:    %v\n", dur2)
	fmt.Printf("   Speedup: %.2fx\n\n", float64(dur1)/float64(dur2))
}

// Value types
type ValueType struct {
	a int64
	b int64
	c int64
}

type PointerType struct {
	a *int64
	b *int64
	c *int64
}

func demoValueTypes() {
	fmt.Println("4. Value Types vs Pointers")

	// Value type - contiguous memory
	values := make([]ValueType, 1000)
	sizeValues := len(values) * int(unsafe.Sizeof(ValueType{}))

	// Pointer type - scattered memory
	pointers := make([]PointerType, 1000)
	sizePointers := len(pointers) * int(unsafe.Sizeof(PointerType{}))
	// Plus actual values on heap

	fmt.Printf("   Value type:  %d bytes\n", sizeValues)
	fmt.Printf("   Pointer type: %d bytes (plus heap)\n", sizePointers)
	fmt.Printf("   Memory saved: %d bytes\n\n", 
		sizePointers-sizeValues+8000) // Rough estimate
}

// Memory pooling
type Document struct {
	ID      string
	Content string
}

var docPool = sync.Pool{
	New: func() interface{} {
		return &Document{}
	},
}

func demoMemoryPooling() {
	fmt.Println("5. Memory Pooling")

	const iterations = 100000

	// Without pool
	start1 := time.Now()
	for i := 0; i < iterations; i++ {
		doc := &Document{
			ID:      fmt.Sprintf("doc-%d", i),
			Content: "content",
		}
		_ = doc
	}
	dur1 := time.Since(start1)

	// With pool
	start2 := time.Now()
	for i := 0; i < iterations; i++ {
		doc := docPool.Get().(*Document)
		doc.ID = fmt.Sprintf("doc-%d", i)
		doc.Content = "content"
		docPool.Put(doc)
	}
	dur2 := time.Since(start2)

	fmt.Printf("   Without pool: %v\n", dur1)
	fmt.Printf("   With pool:    %v\n", dur2)
	fmt.Printf("   Speedup: %.2fx\n\n", float64(dur1)/float64(dur2))
}

// Print memory stats
func printMemStats(phase string) {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)

	fmt.Printf("=== Memory Stats (%s) ===\n", phase)
	fmt.Printf("  Alloc:      %2.1f MB\n", float64(m.Alloc)/1024/1024)
	fmt.Printf("  TotalAlloc: %2.1f MB\n", float64(m.TotalAlloc)/1024/1024)
	fmt.Printf("  Sys:        %2.1f MB\n", float64(m.Sys)/1024/1024)
	fmt.Printf("  NumGC:      %d\n", m.NumGC)
	fmt.Printf("  Goroutines: %d\n", runtime.NumGoroutine())
	fmt.Println()
}

// Optimization tips
func printOptimizationTips() {
	fmt.Println("=== Memory Optimization Tips ===")
	fmt.Println()
	fmt.Println("1. Use value types instead of pointers when possible")
	fmt.Println("   type Doc struct { id int64 }  // Better")
	fmt.Println("   type Doc struct { id *int64 } // Avoid")
	fmt.Println()
	fmt.Println("2. Pre-allocate slices and maps")
	fmt.Println("   make([]int, 0, 1000)  // Better")
	fmt.Println("   var slice []int       // Avoid")
	fmt.Println()
	fmt.Println("3. Use sync.Pool for temporary objects")
	fmt.Println("   Reduces allocations and GC pressure")
	fmt.Println()
	fmt.Println("4. Avoid string concatenation in loops")
	fmt.Println("   Use strings.Builder instead")
	fmt.Println()
	fmt.Println("5. Set GOGC environment variable")
	fmt.Println("   GOGC=100   // Default")
	fmt.Println("   GOGC=50    // Less memory, more CPU")
	fmt.Println("   GOGC=200   // More memory, less CPU")
	fmt.Println()
	fmt.Println("6. Use -ldflags for smaller binary")
	fmt.Println("   -ldflags \"-s -w\"")
	fmt.Println()
	fmt.Println("7. Use UPX for compression")
	fmt.Println("   upx --best binary")
	fmt.Println()
	fmt.Println("8. Build with CGO_ENABLED=0")
	fmt.Println("   Removes C runtime dependencies")
	fmt.Println()
}

// Additional imports for demo
import (
	"strings"
	_ "unsafe"
)
