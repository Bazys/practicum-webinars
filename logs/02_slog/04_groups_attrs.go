// Пример 4: контекст логгера — WithAttrs и WithGroup.
//
// Запуск:
//
//	go run ./02_slog/04_groups_attrs.go
//
// Что демонстрирует:
//   - WithAttrs: добавление общих атрибутов ко всем будущим сообщениям
//   - WithGroup: группировка последующих атрибутов в подобъект JSON
//   - паттерн «общие атрибуты на старте» — request_id, service, version
//   - цепочка WithAttrs/WithGroup не мутирует исходный логгер
package main

import (
	"log/slog"
	"os"
)

func main() {
	base := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// 1) WithAttrs: создаёт НОВЫЙ логгер с общими атрибутами.
	//    ВАЖНО: исходный base не меняется — это безопасно для горутин.
	serviceLog := base.With(
		slog.String("service", "billing"),
		slog.String("version", "1.2.3"),
	)
	serviceLog.Info("сервис запущен")      // service и version будут в каждом сообщении
	serviceLog.Error("ошибка при обработке") // и здесь тоже

	// 2) Цепочка: добавим атрибуты конкретного запроса.
	//    Паттерн HTTP-middleware: request_id добавляем один раз,
	//    дальше все логи в рамках этого запроса несут его с собой.
	reqLog := serviceLog.With(
		slog.String("request_id", "abc-123"),
		slog.String("method", "GET"),
		slog.String("path", "/api/users"),
	)
	reqLog.Info("обработка началась")
	reqLog.Info("обращение к БД", "rows", 42)
	reqLog.Info("ответ отправлен", "status", 200)

	// 3) WithGroup: все последующие атрибуты кладёт в подобъект.
	//    Полезно для структурирования: не плоский JSON, а вложенный.
	userLog := base.WithGroup("user")
	userLog.Info("профиль обновлён",
		slog.Int("id", 42),
		slog.String("email", "x@y.ru"),
	)
	// → {"msg":"профиль обновлён","user":{"id":42,"email":"x@y.ru"}}

	// 4) Группы можно вкладывать друг в друга.
	//    Здесь request > user — два уровня вложенности.
	nested := base.
		WithGroup("request").
		WithGroup("user")
	nested.Info("вложенная группа",
		slog.String("action", "login"),
		slog.Int("id", 7),
	)
	// → {"msg":"...","request":{"user":{"action":"login","id":7}}}
}
