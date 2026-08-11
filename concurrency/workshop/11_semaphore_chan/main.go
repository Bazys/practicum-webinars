// Workshop 11. Semaphore на chan struct{}.
//
// Запуск:
//
//	go run ./11_semaphore_chan
//
// Что показать:
//  1. Семфор = ограничитель параллелизма, построенный только на stdlib.
//  2. Запись в буферизованный канал = «взять слот».
//     Чтение из канала = «вернуть слот».
//  3. В production удобнее golang.org/x/sync/semaphore.Weighted,
//     но понимать «как это сделано» — обязательно.
package main

import (
	"fmt"
	"sync"
	"time"
)

// Semaphore ограничивает число одновременно работающих горутин.
type Semaphore chan struct{}

func NewSemaphore(max int) Semaphore {
	return make(Semaphore, max)
}

func (s Semaphore) Acquire() { s <- struct{}{} }
func (s Semaphore) Release() { <-s }

func main() {
	const maxConcurrent = 3
	sem := NewSemaphore(maxConcurrent)

	var wg sync.WaitGroup
	start := time.Now()

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			sem.Acquire()       // ждём свободный слот
			defer sem.Release() // обязательно возвращаем

			fmt.Printf("-> задача %d стартовала (%.2fs)\n", id, time.Since(start).Seconds())
			time.Sleep(300 * time.Millisecond) // имитация работы
		}(i)
	}

	wg.Wait()
	fmt.Printf("все задачи завершены за %v\n", time.Since(start).Round(time.Millisecond))
	// 10 задач по 300ms, 3 параллельно → ~1200ms (4 «волны»).
}
