// Workshop 08. Fan-Out / Fan-In.
//
// Запуск:
//
//	go run ./08_fan_in_fan_out
//
// Что показать:
//  1. Fan-Out: один источник → N горутин-обработчиков (параллельно).
//  2. Fan-In:  N выходных каналов → один общий канал результатов.
//  3. sync.WaitGroup + goroutine, которая закрывает выход после завершения —
//     стандартная идиома для безопасного закрытия мерж-канала.
package main

import (
	"fmt"
	"sync"
	"time"
)

// gen — источник значений.
func gen(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

// square — одна стадия обработки: читает src, пишет квадраты в новый канал.
func square(src <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range src {
			time.Sleep(50 * time.Millisecond)
			out <- n * n
		}
	}()
	return out
}

// merge — Fan-In: несколько каналов в один.
func merge(chans ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)

	wg.Add(len(chans))
	for _, c := range chans {
		go func(c <-chan int) {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}(c)
	}
	// Закрываем выход только после того, как все источники закрылись.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func main() {
	src := gen(1, 2, 3, 4, 5, 6, 7, 8)

	// Fan-Out: 3 гороутины читают из ОДНОГО src — здесь это просто
	// параллельные обработчики над общей очередью.
	workers := make([]<-chan int, 3)
	for i := range workers {
		workers[i] = square(src)
	}

	// Fan-In: собираем всё обратно.
	for v := range merge(workers...) {
		fmt.Print(v, " ")
	}
	fmt.Println()
}
