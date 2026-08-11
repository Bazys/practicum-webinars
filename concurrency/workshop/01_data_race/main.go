// Workshop 01. Data race — гонка за данными.
//
// Запуск:
//
//	go run ./01_data_race             // результат непредсказуем: 973, 998, 1000...
//	go run -race ./01_data_race       // race detector покажет оба стека
//
// Что показать:
//  1. Несколько прогонов дают разные значения — классический симптом гонки.
//  2. Флаг -race ловит read-modify-write на общей переменной без синхронизации.
//  3. Сравните с вариантом, где счётчик защищён sync/atomic или мьютексом.
package main

import (
	"fmt"
	"sync"
)

func main() {
	var (
		wg    sync.WaitGroup
		count int
	)

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count++ // ← data race: read-modify-write не атомарно
		}()
	}
	wg.Wait()
	fmt.Println("count =", count)
}
