# Postgres в Go и тестирование БД — воркшоп

Рабочий пример к вебинару курса «Продвинутый Go-разработчик». Каталог дополняет
презентацию темами, которые важны на практике, но не уместились на слайдах:
пул соединений, плейсхолдеры Postgres, контексты и таймауты, транзакции с
`SELECT FOR UPDATE`, типы PG (`uuid`/`jsonb`), миграции и три уровня тестов.

Стек: **pgx/v5 (нативный `pgxpool`)** + **golang-migrate** + **testcontainers-go**.

## Что понадобится

- Go 1.23+
- Docker (только для интеграционных тестов и ручного запуска БД)

```sh
make test              # unit-тесты, Docker НЕ нужен
make test-integration  # интеграционные тесты, Docker нужен
```

## Структура примера

```
videos/
├── migrations/                    # миграции (embed'ятся в бинарник)
│   ├── 0001_init.up.sql           #   CREATE TABLE videos (uuid, jsonb, ...)
│   └── 0001_init.down.sql
├── model.go                       # доменная модель Video (uuid + jsonb)
├── repository.go                  # интерфейс VideoRepository + pgx-реализация
├── tx_example.go                  # AddView через транзакцию и SELECT FOR UPDATE
├── handler.go                     # HTTP-слой на интерфейсе репозитория
├── main.go                        # pgxpool + конфиг пула + миграции + graceful shutdown
├── migrate.go                     # накат миграций из embed.FS
├── handler_test.go                # UNIT: мок доменного интерфейса
├── repository_pgxmock_test.go     # UNIT: мок драйвера pgx (проверка SQL)
└── repository_integration_test.go //go:build integration  — testcontainers-go
```

## Сценарий воркшопа по шагам

Каждый шаг — отдельная тема. Показывайте файл, прогоняйте команду, обсуждайте.

### Шаг 1. Подключение к Postgres: почему pgx, а не lib/pq
- **Слайд 10.** `lib/pq` больше не развивается, `pgx` — стандарт.
- Смотрим `main.go`: `newPool()` создаёт `pgxpool.Pool` (нативный пул pgx, а не `database/sql`).
- Главное преимущество: pgx работает с типами Postgres напрямую, без потерь через `database/sql`.

### Шаг 2. Пул соединений (нет в презентации)
- `main.go`, функция `newPool`: `MaxConns`, `MinConns`, `MaxConnLifetime`, `MaxConnIdleTime`.
- Обсудить: что будет, если лимиты не задать (исчерпание соединений в БД под нагрузкой).

### Шаг 3. Плейсхолдеры: `$1` вместо `?`
- **Ловушка со слайдов 8–9.** В Postgres позиционные плейсхолдеры — `$1, $2, …`,
  а не `?` (как в MySQL/SQLite).
- Смотрим `repository.go`, `List`: `... LIMIT $1`.
- **Демо «сломанного плейсхолдера»**: временно заменить `$1` на `?`, показать ошибку
  синтаксиса, вернуть обратно. Старая версия `main.go` в git как раз содержала этот баг.

### Шаг 4. Контекст и таймауты (нет в презентации)
- `handler.go`: каждый хендлер делает `context.WithTimeout(r.Context(), queryTimeout)`
  и передаёт `ctx` в репозиторий, а оттуда — в pgx.
- Обсудить: без дедлайна зависший запрос держит соединение из пула бесконечно.

### Шаг 5. Типы Postgres: `uuid` и `jsonb` (нет в презентации)
- `model.go`: `pgtype.UUID` ↔ колонка `uuid`, `json.RawMessage` ↔ `jsonb`.
- `migrations/0001_init.up.sql`: `gen_random_uuid()`, `jsonb`, `timestamptz`.
- `repository.go`, `Create`: передаём `MetadataFields` (структуру), pgx сам сериализует в jsonb.

### Шаг 6. Транзакции и `SELECT FOR UPDATE` (нет в презентации)
- `tx_example.go`: `AddViewInTransaction` — шаблон «прочитал-заблокировал-записал».
  `SELECT ... FOR UPDATE` блокирует строку до коммита.
- `repository.go`: `AddView` — атомарный `UPDATE views = views + 1`. Для счётчика это
  проще и быстрее. Транзакция показана как универсальный приём для случаев, когда
  новое значение нельзя выразить одним SQL.

### Шаг 7. Миграции (нет в презентации)
- `migrate.go`: миграции встраиваются в бинарник через `//go:embed`, накатываются
  через `golang-migrate`. Бинарник самодостаточен.
- `main.go`: `runMigrations(ctx, dsn)` при старте.

### Шаг 8. Тесты. Уровень 1 — мок доменного интерфейса
- **Слайд 13.** `handler_test.go`: `mockVideoRepository` мокает `VideoRepository`.
- Тестируем логику handler: доменная ошибка `ErrVideoNotFound` → HTTP 404,
  техническая ошибка → 500. SQL не трогаем, БД не нужна.
```sh
go test -run TestAPI -v
```

### Шаг 9. Тесты. Уровень 2 — мок драйвера (pgxmock)
- `repository_pgxmock_test.go`: мокаем сам pgx и проверяем **точный SQL** и
  плейсхолдеры (`$1`), порядок аргументов, обработку `pgx.ErrNoRows`, код `23505`.
```sh
go test -run TestRepository -v
```

### Шаг 10. Тесты. Уровень 3 — интеграционные с testcontainers-go
- **Слайды 14–16.** `repository_integration_test.go` (`//go:build integration`):
  тест **сам поднимает Postgres в Docker**, накатывает миграции, гоняет запросы,
  убирает контейнер за собой.
- Ключевой тест — гонка счётчика: 100 конкурентных `AddView` должны дать ровно 100.
```sh
make test-integration
```

## Что показать сверх презентации (шпаргалка ведущему)

| Тема | Где | Почему важно |
|---|---|---|
| Пул соединений | `main.go: newPool` | без лимитов БД падает под нагрузкой |
| `$1` vs `?` | `repository.go` | частый баг при переходе с MySQL |
| Контекст/таймаут | `handler.go` | защита пула от зависших запросов |
| `uuid`, `jsonb` | `model.go`, миграции | типы PG вне стандарта SQL |
| `SELECT FOR UPDATE` | `tx_example.go` | корректность при конкурентных запросах |
| Миграции (`embed`) | `migrate.go` | самодостаточный бинарник |
| testcontainers-go | `*_integration_test.go` | тесты с реальной БД без ручной настройки |

## Сноска про sqlx и sqlc

В презентации упоминаются `jmoiron/sqlx` (NamedQuery/NamedExec, слайды 9, 11) и
в корне репозитория лежит `sqlc.yaml`. В этом примере они не используются — выбран
нативный pgx. Идея для обсуждения: sqlx удобен поверх `database/sql`, но pgx
повторяет его эргономику (`pgx.RowToStructByName`) и при этом быстрее и «ближе» к
Postgres. sqlc же генерирует типобезопасный код из SQL — это следующий шаг после
рукописных запросов, его стоит показать отдельным примером.
