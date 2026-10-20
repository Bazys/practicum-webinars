package benchmark

import "testing"

// Result makes the constructed slice observable to the benchmark caller.
var Result []int

const size = 1_000_000

func BenchmarkBuild(b *testing.B) {
	b.Run("Grow", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			// A nil slice deliberately matches the original growth strategy.
			var values []int
			for i := 0; i < size; i++ {
				values = append(values, i)
			}
			Result = values
		}
	})
	b.Run("Prealloc", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			values := make([]int, 0, size)
			for i := 0; i < size; i++ {
				values = append(values, i)
			}
			Result = values
		}
	})
}
