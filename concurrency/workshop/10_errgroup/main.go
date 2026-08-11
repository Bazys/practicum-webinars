// Workshop 10. errgroup — структурированная конкурентность.
//
// Запуск:
//
//	go run ./10_errgroup
//
// Зависимости: golang.org/x/sync/errgroup (см. go.mod в этой папке).
//
// Что показать:
//
//  1. errgroup.WithContext возвращает контекст, который отменяется автоматически
//     при возврате первой ошибки из любой гороутины группы.
//
//  2. g.Wait() возвращает первую (с точки зрения возвращённой) ошибку.
//
//  3. g.SetLimit(N) — бонус: ограничивает число одновременно работающих горутин
//     (по сути встроенный семафор).
//
//     вариант реализации pipeline с помощью errgroup
//     https://pkg.go.dev/golang.org/x/sync/errgroup#example-Group-Pipeline
package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/sync/errgroup"
)

func fetch(ctx context.Context, url string) error {
	// Имитация HTTP-запроса с возможной ошибкой.
	delay := time.Duration(50+rand.Intn(150)) * time.Millisecond
	select {
	case <-time.After(delay):
		if rand.Intn(3) == 0 { // ~1/3 запросов «падает»
			return fmt.Errorf("fetch %s: internal server error", url)
		}
		fmt.Printf("  ✓ %s (%v)\n", url, delay)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func fetchAll(urls []string) error {
	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(3) // не больше 3 одновременных запросов

	for _, url := range urls {
		url := url // capture для гороутины (до Go 1.22)
		g.Go(func() error {
			return fetch(ctx, url)
		})
	}
	// Ждём всех. Первая ошибка отменит остальных через ctx.
	return g.Wait()
}

func main() {
	rand.Seed(time.Now().UnixNano())

	urls := []string{
		"https://api.example.com/users",
		"https://api.example.com/orders",
		"https://api.example.com/products",
		"https://api.example.com/payments",
		"https://api.example.com/inventory",
		"https://api.example.com/reports",
	}

	if err := fetchAll(urls); err != nil {
		if errors.Is(err, context.Canceled) {
			fmt.Println("отменено контекстом:", err)
		} else {
			fmt.Println("первая ошибка группы:", err)
		}
		return
	}
	fmt.Println("все запросы успешны")
}
