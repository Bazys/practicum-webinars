.PHONY: gomod lint test test-integration run migrate-up docker-up docker-down

# Управление зависимостями.
gomod:
	@echo ">> go mod tidy"
	go mod tidy

# Линтер (нужен установленный golangci-lint).
lint:
	@echo ">> golangci-lint"
	golangci-lint run

# Unit-тесты: быстро, БЕЗ Docker. Моки интерфейса + pgxmock.
test:
	@echo ">> unit tests"
	@go clean -testcache
	go test -race -count=1 -cover ./...

# Интеграционные тесты: поднимают Postgres в Docker через testcontainers-go.
# Нужен запущенный Docker-демон.
test-integration:
	@echo ">> integration tests (Docker required)"
	@go clean -testcache
	go test -tags=integration -race -count=1 -timeout=180s ./...

# Запуск приложения. Подразумевает, что Postgres уже поднят (см. docker-up).
# Миграции накатываются автоматически при старте.
run:
	go run .

# Поднять локальный Postgres (как на слайде docker-compose).
up:
	docker compose up -d

down:
	docker compose down
