# Пример: интеграция `slog` + `os` + `io`

Финальный пример вебинара — соединяет все три пакета в типичный production-сетап.

## Запуск

```bash
cd examples
go run ./04_integration/log_to_file.go
```

Запустите 2–3 раза подряд — увидите, как лог накапливается в `app.log`
(благодаря `O_APPEND`):

```bash
go run ./04_integration/log_to_file.go
go run ./04_integration/log_to_file.go
cat app.log
```

После проверки можно удалить лог:
```bash
rm app.log
```

## Что делает пример

В 30 строках кода соединяет всё, пройденное на вебинаре:

| Шаг | Пакет | Что |
|---|---|---|
| 1 | `os` | `os.OpenFile("app.log", O_APPEND\|O_CREATE\|O_WRONLY, 0644)` — открыть лог в режиме append |
| 2 | `io` | `io.MultiWriter(os.Stdout, logFile)` — дублировать вывод в консоль и файл |
| 3 | `log/slog` | `slog.NewJSONHandler(multi, opts)` — структура и формат |
| 4 | `log/slog` | `slog.SetDefault(logger)` — сделать глобальным |
| 5 | `log/slog` | `logger.With(service, version)` — общие атрибуты на старте |
| 6 | `os` | `logFile.Sync()` — сброс буферов перед выходом |

## Вывод

В терминале и в `app.log` появляются строки вида:

```json
{"time":"2026-07-09T23:30:00.123+03:00","level":"INFO","msg":"сервис запущен","service":"myapp","version":"1.0.0","port":8080}
{"time":"...","level":"INFO","msg":"обработан запрос","service":"myapp","version":"1.0.0","user_id":42,"method":"GET","path":"/api/users","status":200}
{"time":"...","level":"ERROR","msg":"не удалось отправить email","service":"myapp","version":"1.0.0","err":"smtp timeout","to":"user@example.com"}
```

Каждая запись содержит:
- `time`, `level`, `msg` — от `slog`
- `service`, `version` — общие атрибуты (из `WithAttrs`)
- остальные поля — из конкретного вызова

## Production-расширения

В реальном проекте обычно добавляют:

1. **Ротацию логов** через `gopkg.in/natefinch/lumberjack.v2`:
   ```go
   &lumberjack.Logger{Filename: "app.log", MaxSize: 100, MaxBackups: 7}
   ```
   `lumberjack.Logger` реализует `io.Writer` — вставляется вместо `logFile`.

2. **Контекст запроса** — `request_id` через middleware:
   ```go
   logger.With("request_id", reqID)
   ```

3. **В современном production (K8s/Docker)** — писать только в `os.Stdout`,
   а сборщик логов (fluentd, vector, promtail) забирает и отправляет в Loki/ELK.
   Файловый случай (этот пример) — база для понимания.

## Связанные слайды

29, 30 в `slides/webinar.md`.

## Связанный конспект

`notes/04-integration.md`.
