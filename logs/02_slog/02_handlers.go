// Пример 2: TextHandler vs JSONHandler — два формата вывода.
//
// Запуск:
//
//	go run ./02_slog/02_handlers.go
//
// Что демонстрирует:
//   - создание логгера с разными хендлерами через slog.New()
//   - сравнение вывода: Text (человекочитаемый) vs JSON (для машин)
//   - как переключить глобальный логгер через slog.SetDefault
package main

import (
	"log/slog"
	"os"
)

func main() {
	// Общие данные для обеих демонстраций.
	logSomething := func(l *slog.Logger) {
		l.Info("пользователь вошёл",
			slog.String("user", "alice"),
			slog.Int("user_id", 42),
			slog.String("ip", "10.0.0.1"),
		)
		l.Warn("мало памяти", "free_mb", 100)
	}

	// 1) TextHandler — человекочитаемый формат key=value.
	//    Подходит для dev-режима и терминала.
	textLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	println("--- TextHandler (dev) ---")
	logSomething(textLogger)

	// 2) JSONHandler — машиночитаемый JSON.
	//    Подходит для production: ELK, Loki, Datadog.
	jsonLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	println("\n--- JSONHandler (prod) ---")
	logSomething(jsonLogger)

	// 3) slog.SetDefault — делает логгер глобальным.
	//    После этого slog.Info(...) использует именно его.
	slog.SetDefault(jsonLogger)
	println("\n--- после SetDefault(jsonLogger): ---")
	slog.Info("это сообщение через slog.Info", "source", "default")
}
