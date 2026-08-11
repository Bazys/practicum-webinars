// Workshop 05. select — тайм-ауты, default, отмена.
//
// Запуск:
//
//	go run ./05_select_timeout
//
// Что показать:
//  1. select ждёт первый готовый case (если их несколько — выбирает случайно).
//  2. time.After — простой способ сделать тайм-аут.
//  3. default делает select неблокирующим.
//  4. nil-канал в select — никогда не готов (удобно для динамического вкл/выкл case).
package main

import (
	"fmt"
	"time"
)

func slowProducer() <-chan string {
	ch := make(chan string)
	go func() {
		time.Sleep(300 * time.Millisecond)
		ch <- "done"
	}()
	return ch
}

func main() {
	// 1) Тайм-аут на чтение.
	ch := slowProducer()
	select {
	case v := <-ch:
		fmt.Println("получено:", v)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("timeout: продюсер не успел за 100ms")
	}

	// 2) default — неблокирующее чтение.
	empty := make(chan int, 1)
	select {
	case v := <-empty:
		fmt.Println("взяли:", v)
	default:
		fmt.Println("пусто, идём дальше без блокировки")
	}

	// 3) Цикл с тайм-аутом — полезен, когда хочется мягко прервать ожидание.
	deadline := time.After(250 * time.Millisecond)
	ch2 := slowProducer()
	select {
	case v := <-ch2:
		fmt.Println("успели до дедлайна:", v)
	case <-deadline:
		fmt.Println("дедлайн пришёл раньше")
	}
}
