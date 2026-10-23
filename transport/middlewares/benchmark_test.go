package middlewares

import (
	"runtime/debug"
	"testing"
)

// BenchmarkStackFrames documents the cost of the structured stack capture
// used by the Recovery middleware. It runs only on the panic path.
//
// The structured capture is cheaper than debug.Stack, which formats the whole
// trace into a single string.
func BenchmarkStackFrames(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = stackFrames(1)
	}
}

// BenchmarkDebugStack is the reference of the previous implementation.
func BenchmarkDebugStack(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = debug.Stack()
	}
}
