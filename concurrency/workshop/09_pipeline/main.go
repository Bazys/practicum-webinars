// Workshop 09. Pipeline — цепочка стадий через каналы.
//
// Запуск:
//
//	go run ./09_pipeline
//
// Что показать:
//  1. Каждая стадия — функция, возвращающая канал.
//  2. Главный гороутин читает из последнего канала и сам выходит при close.
//  3. defer close(out) — обязательная дисциплина: если вход закрылся,
//     выход тоже должен закрыться, иначе consumer зависнет навсегда.
//  4. Для отказоустойчивости в реальном коде сюда добавляют context.
//     Статья про pipeline в "The Go Blog" https://go.dev/blog/pipelines
package main

import "fmt"

// Стадия 1: источник.
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

// Стадия 2: возведение в квадрат.
func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

// Стадия 3: фильтр (только чётные).
func filterEven(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			if n%2 == 0 {
				out <- n
			}
		}
	}()
	return out
}

func main() {
	// gen → square → filterEven → print
	pipeline := filterEven(square(gen(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)))

	for v := range pipeline {
		fmt.Print(v, " ")
	}
	fmt.Println()
	// Вывод: квадраты 1..10 = 1 4 9 ... 100, чётные из них → 4 16 36 64 100
}
