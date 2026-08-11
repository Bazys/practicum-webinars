package main

import "testing"

// BenchmarkMutex — параллельные чтения под эксклюзивной блокировкой.
func BenchmarkMutex(b *testing.B) {
	c := &CacheMutex{data: makeData()}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = c.Get("1")
		}
	})
}

// BenchmarkRWMutex — параллельные чтения под RLock (несколько читателей).
func BenchmarkRWMutex(b *testing.B) {
	c := &CacheRWMutex{data: makeData()}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = c.Get("1")
		}
	})
}
