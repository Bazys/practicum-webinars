// Workshop 07. Worker Pool.
//
// Запуск:
//
//	go run ./07_worker_pool
//
// Что показать:
//  1. Фиксированное число воркеров = ограничение параллелизма.
//     Полезно для внешних вызовов (БД, HTTP), где нельзя «положить» сервис.
//  2. Канал jobs раздаёт задачи, results — собирает ответы.
//  3. close(jobs) — сигнал «работы больше нет», воркеры выходят из range.
//  4. Количество воркеров не равно GOMAXPROCS: IO-bound → больше, CPU-bound ≈ NumCPU.
package main

import (
	"fmt"
	"sync"
	"time"
)

const numWorkers = 3

func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs {
		fmt.Printf("worker %d: начал задачу %d\n", id, j)
		time.Sleep(200 * time.Millisecond) // имитация работы
		results <- j * j
	}
}

func main() {
	jobs := make(chan int, 10)
	results := make(chan int, 10)

	var wg sync.WaitGroup
	// Запускаем пул воркеров.
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg)
	}

	// Раздаём задачи.
	for j := 1; j <= 9; j++ {
		jobs <- j
	}
	close(jobs) // работу больше не принимаем

	// Закрываем results после завершения всех воркеров.
	go func() {
		wg.Wait()
		close(results)
	}()

	// Собираем результаты (порядок НЕ гарантирован!).
	sum := 0
	for r := range results {
		sum += r
	}
	fmt.Println("сумма квадратов 1..9 =", sum) // 285
}
