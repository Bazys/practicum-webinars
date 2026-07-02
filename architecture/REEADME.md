# HOW TO USE

Запускаем сервер:

```sh
go run main.go
```

Проверяем успешный кейс (Alice делает заказ):

```sh
curl -X POST http://localhost:8888/orders \
-H "Content-Type: application/json" \
-d '{"user_id":"user-1", "product_id":"prod-1", "quantity":2}'
```

Ожидаемый ответ: {"id":"order-1","user_id":"user-1","product_id":"prod-1","quantity":2,"total":200,"status":"paid"}

Проверяем кейс с ошибкой биллинга (Bob делает заказ):

```sh
curl -X POST http://localhost:8888/orders \
-H "Content-Type: application/json" \
-d '{"user_id":"user-2", "product_id":"prod-2", "quantity":1}'
```

Ожидаемый ответ: billing failed: insufficient funds

## проблемы этого кода

### Невозможность тестирования

Как нам протестировать расчет total = price * float64(req.Quantity) без поднятия HTTP-сервера?
Никак. Функция намертво привязана к http.ResponseWriter.

### Смешение ошибок

HTTP статусы (StatusPaymentRequired) формируются на основе бизнес-правил. Если завтра мы захотим добавить gRPC, нам придется переписывать эту логику.

### Скрытые зависимости

Функция createOrderHandler зависит от глобальных переменных usersDB, productsDB. Это делает её непредсказуемой.
Нарушение Single Responsibility Principle (SRP): Хендлер занимается всем: парсингом, бизнес-логикой, БД и вызовом внешних API.

## Начнем резать этот монолит

Первым делом мы выделим бизнес-правила (домен) и вынесем его в отдельный пакет internal/domain, чтобы он перестал знать о существовании net/http и баз данных.


### Выделение Домена и Сервиса

```text
webinar-architecture/
├── go.mod
├── main.go                 <-- Теперь это точка входа и адаптеры
├── internal/
│   ├── domain/
│   │   └── order.go        <-- Чистая бизнес-сущность
│   └── service/
│       └── order_service.go <-- Бизнес-логика (Use Case) + ИНТЕРФЕЙСЫ
```

На текущем этапе наши моки-адаптеры лежат прямо в main.go.
В реальных проектах их много, и main раздувается.
Кроме того, мы захотим использовать настоящую PostgreSQL.
На следующем шаге мы выделим Инфраструктурный слой
(пакеты repository и integration) и посмотрим,
как internal защищает наш домен от внешнего мира.

## Финальная сборка Clean Architecture

```text
webinar-architecture/
├── go.mod
├── main.go                           <-- ТОЛЬКО сборка зависимостей (Composition Root)
└── internal/
    ├── domain/
    │   └── order.go                  <-- (Без изменений)
    ├── service/
    │   └── order_service.go          <-- (Без изменений)
    ├── repository/
    │   └── memory/
    │       ├── user.go               <-- Адаптер БД пользователей
    │       ├── product.go            <-- Адаптер БД товаров
    │       └── order.go              <-- Адаптер БД заказов
    ├── integration/
    │   └── billing.go                <-- Адаптер внешнего сервиса биллинга
    └── transport/
        └── http/
            └── handler.go            <-- HTTP хендлеры
```

Мы создаем пакет memory, чтобы показать,
что это конкретная реализация (in-memory).
Завтра мы напишем пакет internal/repository/postgres.
