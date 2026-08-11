// Workshop 06. context.Context — отмена и дедлайны.
//
// Запуск:
//
//	go run ./06_context_cancel
//
// Что показать:
//  1. context.WithCancel — ручная отмена (например, по сигналу от клиента).
//  2. context.WithTimeout — автоматическая отмена по таймеру.
//  3. ctx.Err() отличает «отменили» от «дедлайн прошёл».
//  4. ВСЕГДА вызываем cancel() (через defer), иначе утечём.
//  5. context values — пользоваться осторожно, не как глобалом.
package main

import (
	"context"
	"fmt"
	"time"
)

// worker имитирует долгую работу, регулярно проверяя отмену.
func worker(ctx context.Context, id int) error {
	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			fmt.Printf("worker %d: итерация %d\n", id, i)
			if i >= 10 {
				return nil
			}
		}
	}
}

func main() {
	// 1) Ручная отмена через 350ms.
	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()

	err := worker(ctx, 1)
	fmt.Println("worker 1 завершился с:", err)
	// → context.DeadlineExceeded

	// 2) Передаём значение через контекст (только request-scoped!).
	ctx2 := context.WithValue(context.Background(), "requestID", "abc-123")
	if v, ok := ctx2.Value("requestID").(string); ok {
		fmt.Println("requestID из контекста:", v)
	}

	// 3) Явная отмена по условию.
	ctx3, cancel3 := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel3() // приняли решение остановить
	}()
	fmt.Println("worker 3:", worker(ctx3, 3))
}
