// Пример 1: базовые вызовы slog и уровни логирования.
//
// Запуск:
//
//	go run ./02_slog/01_basic.go
//
// Что демонстрирует:
//   - slog.Debug/Info/Warn/Error через стандартный логгер
//   - два способа передачи атрибутов: сокращённый (ключ-значение) и явный
//   - типизированные конструкторы атрибутов: slog.String/Int/Bool/Duration
//   - что Debug по умолчанию скрыт (уровень по умолчанию = Info)
package main

import (
	"log/slog"
	"time"
)

func main() {
	// Стандартный логгер: TextHandler на os.Stderr, уровень Info.
	// Debug будет скрыт — увидим его в примере 03_levels_source.go,
	// когда повысим уровень через HandlerOptions.

	// 1) Уровни логирования.
	slog.Debug("это сообщение скрыто")          // уровень < Info
	slog.Info("сервер запущен", "port", 8080)    // видно
	slog.Warn("мало памяти", "free_mb", 100)     // видно
	slog.Error("ошибка БД", "err", "connection refused") // видно

	// 2) Сокращённая форма: чередуем "ключ, значение".
	//    Ключ ОБЯЗАТЕЛЬНО строка, иначе паника.
	slog.Info("login", "user", 42, "ip", "10.0.0.1", "success", true)

	// 3) Явная форма через типизированные атрибуты — лучше для production,
	//    потому что сохраняет тип значения в JSON.
	slog.Info("login",
		slog.Int("user", 42),
		slog.String("ip", "10.0.0.1"),
		slog.Bool("success", true),
		slog.Duration("elapsed", 150*time.Millisecond),
		slog.Time("at", time.Now()),
	)

	// 4) slog.Any — универсальный конструктор для любых типов.
	//    Для своих типов вызывает LogValuer, если реализован (см. пример 05).
	type Request struct{ Method, Path string }
	slog.Info("запрос получен", "req", Request{Method: "GET", Path: "/users"})
}
