// Пример 3: HandlerOptions — уровень логирования и AddSource.
//
// Запуск:
//
//	go run ./02_slog/03_levels_source.go
//
// Что демонстрирует:
//   - настройка минимального уровня через HandlerOptions.Level
//   - включение Debug-сообщений (по умолчанию скрыты)
//   - AddSource: true — добавляет поле с файлом и строкой вызова
//   - сравнение разных уровней в действии
package main

import (
	"log/slog"
	"os"
)

func main() {
	// 1) HandlerOptions с LevelDebug и AddSource.
	//    LevelDebug включает ВСЕ уровни (Debug и выше).
	//    AddSource добавляет поле source = {function, file, line}.
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))

	// Теперь Debug виден.
	logger.Debug("отладочное сообщение теперь отображается")
	logger.Info("обычное info")
	logger.Warn("предупреждение")
	logger.Error("ошибка")

	// 2) Сравнение уровней. Создадим несколько логгеров с разными порогами.
	levels := []struct {
		name  string
		level slog.Leveler
	}{
		{"LevelError (только ошибки)", slog.LevelError},
		{"LevelWarn", slog.LevelWarn},
		{"LevelInfo (по умолчанию)", slog.LevelInfo},
		{"LevelDebug (всё)", slog.LevelDebug},
	}

	for _, l := range levels {
		opts := &slog.HandlerOptions{Level: l.level}
		log := slog.New(slog.NewTextHandler(os.Stdout, opts))
		log.Info("инфо при уровне "+l.name, "level", l.name)
		log.Debug("дэбаг при уровне "+l.name, "level", l.name)
	}

	// 3) Замечание про накладные расходы AddSource:
	//    определение стека вызовов — относительно дорогая операция.
	//    В hot path production обычно выключают, оставляя только для Error.
}
