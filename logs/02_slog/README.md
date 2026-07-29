# Примеры: пакет `log/slog`

Структурированное логирование (стандарт Go 1.21+).

## Запуск

```bash
cd examples
go run ./02_slog/01_basic.go
go run ./02_slog/02_handlers.go
go run ./02_slog/03_levels_source.go
go run ./02_slog/04_groups_attrs.go
go run ./02_slog/05_logvaluer.go
```

> Каждый файл — отдельный `package main`. Запускайте по одному.

## Что в каждом файле

### `01_basic.go` — уровни и атрибуты
- `slog.Debug/Info/Warn/Error`
- два способа передачи атрибутов: сокращённый и через типизированные конструкторы
- `slog.Duration`, `slog.Time`, `slog.Any`

### `02_handlers.go` — TextHandler vs JSONHandler
- сравнение форматов вывода
- `slog.SetDefault` — переключение глобального логгера

### `03_levels_source.go` — HandlerOptions
- `Level: slog.LevelDebug` — включение Debug
- `AddSource: true` — поле `source` с файлом и строкой

### `04_groups_attrs.go` — контекст логгера
- `WithAttrs` — общие атрибуты (service, version, request_id)
- `WithGroup` — вложенные группы в JSON
- цепочка не мутирует исходный логгер

### `05_logvaluer.go` — кастомное форматирование типа
- интерфейс `slog.LogValuer`
- скрытие пароля, маскирование email и номера карты
- безопасность «по умолчанию»

## Подводные камни

- **Сокращённая форма паникует, если ключ не строка.** `slog.Info("x", 1, 2)` → panic.
- **`AddSource: true` замедляет логирование** — определение стека дорогое.
- **`slog.Default()` — глобальное состояние.** В тестах создавайте локальный логгер.
- **`WithAttrs` создаёт новый логгер** — не злоупотребляйте в циклах (аллокации).

## Когда выбирать zap/zerolog вместо slog

- microbenchmarks показывают, что логирование — узкое место (редко)
- нужна совместимость со старой кодовой базой

В 90% случаев **slog достаточно**: стандартная библиотека, общая абстракция,
интегрируется со сторонними хендлерами (`slog-zap` и т.п.).

## Связанные слайды

12, 13, 14, 15, 16, 17, 18, 19 в `slides/webinar.md`.

## Связанный конспект

`notes/02-slog.md`.
