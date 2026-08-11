// Пример интеграции: slog + os.OpenFile + io.MultiWriter.
//
// Это финальный пример вебинара — он объединяет ВСЕ три пакета:
//   - log/slog  — структурированное логирование
//   - os        — открытие файла в режиме append
//   - io        — MultiWriter для дублирования вывода
//
// Запуск:
//
//	go run ./04_integration/log_to_file.go
//
// После запуска:
//   - JSON-лог появится в терминале (os.Stdout)
//   - и в файле app.log рядом (допишется, если файл уже был)
//
// Запустите 2-3 раза подряд, чтобы увидеть, как лог накапливается в app.log.
package main

import (
	"io"
	"log"
	"log/slog"
	"os"
)

func main() {
	// 1) Открываем лог-файл в режиме ДОЗАПИСИ.
	//    O_APPEND  — писать в конец
	//    O_CREATE  — создать, если не существует
	//    O_WRONLY  — только для записи
	//    0644      — права: владелец rw, остальные r
	logFile, err := os.OpenFile("app.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("не могу открыть лог-файл: %v", err)
	}
	defer logFile.Close()

	// 2) io.MultiWriter возвращает Writer, который дублирует запись
	//    во все переданные приёмники. Здесь: и в консоль, и в файл.
	//    Это позволяет видеть лог в реальном времени при разработке
	//    и одновременно сохранять его для анализа.
	multi := io.MultiWriter(os.Stdout, logFile)

	// 3) Создаём slog-логгер с JSONHandler — для production.
	//    HandlerOptions.Level задаёт минимальный уровень (Info по умолчанию).
	//    Для dev можно использовать NewTextHandler вместо NewJSONHandler.
	logger := slog.New(slog.NewJSONHandler(multi, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// 4) Делаем этот логгер глобальным.
	//    После этого slog.Info(...) использует именно его.
	slog.SetDefault(logger)

	// 5) Добавляем общие атрибуты — они появятся в КАЖДОЙ записи.
	//    Паттерн: service/version/request_id задаём один раз на старте.
	logger = logger.With(
		slog.String("service", "myapp"),
		slog.String("version", "1.0.0"),
	)
	slog.SetDefault(logger)

	// === Всё настроено, теперь просто пишем логи как обычно ===

	slog.Info("сервис запущен", "port", 8080)
	slog.Info("подключение к БД установлено", "host", "localhost", "db", "myapp")

	// Логируем запрос — атрибуты типизированы.
	userID := 42
	method := "GET"
	path := "/api/users"
	slog.Info("обработан запрос",
		slog.Int("user_id", userID),
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status", 200),
	)

	// Предупреждение и ошибка.
	slog.Warn("медленный запрос", "path", "/api/search", "ms", 1500)
	slog.Error("не удалось отправить email", "err", "smtp timeout", "to", "user@example.com")

	// 6) Перед выходом сбрасываем буферы на диск.
	//    Sync — относительно дорогая операция, в горячем пути её не делают,
	//    но при graceful shutdown стоит вызвать, чтобы не потерять последние записи.
	if err := logFile.Sync(); err != nil {
		slog.Error("Sync лог-файла", "err", err)
	}

	slog.Info("сервис завершил работу")
}
