// Workshop 12. Goroutine leak — было/стало.
//
// Запуск:
//
//	go run ./12_goroutine_leak
//
// Что показать:
//  1. fetchBad запускает гороутину, которая пишет в небуферизованный канал.
//     Если читатель отвалился по тайм-ауту — отправитель повиснет НАВСЕГДА.
//  2. fetchGood использует буфер = 1 + context: гороутин гарантированно завершается,
//     даже если результат уже никому не нужен.
//  3. runtime.NumGoroutine() позволяет поймать утечку в рантайме.
package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// ❌ Утечка: отправитель ждёт читателя, которого уже нет.
func fetchBad() <-chan int {
	ch := make(chan int) // небуферизованный
	go func() {
		time.Sleep(50 * time.Millisecond)
		ch <- 42 // повиснет тут, если читатель ушёл
		fmt.Println("  (bad) гороутина завершилась — НО в утечке этого не видно")
	}()
	return ch
}

// ✅ Безопасно: буфер = 1 + проброс отмены через ctx.
func fetchGood(ctx context.Context) (int, error) {
	ch := make(chan int, 1) // буфер позволяет отправителю не блокироваться
	go func() {
		time.Sleep(50 * time.Millisecond)
		ch <- 42
	}()

	select {
	case v := <-ch:
		return v, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func main() {
	before := runtime.NumGoroutine()

	// 1) ПЛОХО: читатель отваливается по тайм-ауту.
	badCh := fetchBad()
	select {
	case <-badCh:
		fmt.Println("bad: успели прочитать")
	case <-time.After(10 * time.Millisecond):
		fmt.Println("bad: тайм-аут, читатель ушёл → отправитель утечён")
	}

	// 2) ХОРОШО: отмена пробрасывается, ничего не виснет.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	_, err := fetchGood(ctx)
	fmt.Println("good: результат err =", err)
	cancel()

	time.Sleep(100 * time.Millisecond) // даём утечке «проявиться»
	after := runtime.NumGoroutine()
	fmt.Printf("гороутины: было=%d стало=%d (дельта=%d)\n", before, after, after-before)
}
