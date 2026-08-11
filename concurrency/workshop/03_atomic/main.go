// Workshop 03. sync/atomic — lock-free счётчики и CAS.
//
// Запуск:
//
//	go run ./03_atomic
//	go run -race ./03_atomic
//
// Что показать:
//  1. atomic.Int64.Add даёт честный счёт без блокировок.
//  2. CompareAndSwap — основа lock-free алгоритмов.
//  3. Типы-обёртки (atomic.Int64/Bool/Pointer[T]) появились в Go 1.19 —
//     ими пользоваться проще и безопаснее, чем сырыми функциями atomic.AddInt64.
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	// 1) Простой счётчик
	var counter atomic.Int64
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Add(1)
		}()
	}
	wg.Wait()
	fmt.Println("counter:", counter.Load()) // всегда 1000

	// 2) CompareAndSwap (CAS) — безопасно обновить значение только если оно не изменилось.
	var state atomic.Int64
	state.Store(1)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				old := state.Load()
				new := old + 1
				if state.CompareAndSwap(old, new) {
					break // успешно обновили
				}
				// иначе — кто-то успел раньше, пробуем снова
			}
		}()
	}
	wg.Wait()
	fmt.Println("state (CAS):", state.Load())

	// 3) atomic.Bool — удобно как флаг «выполняется»/«остановлено».
	var stopped atomic.Bool
	stopped.Store(true)
	fmt.Println("stopped:", stopped.Load())
}
