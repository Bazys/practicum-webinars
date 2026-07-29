//go:build integration

// Интеграционные тесты поднимают НАСТОЯЩУЮ PostgreSQL в Docker-контейнере.
// Сборочный тег integration гарантирует, что `go test ./...` по умолчанию
// их НЕ запускает — нужен явный `go test -tags=integration`.
//
// Запуск:
//
//	go test -tags=integration -v -run TestRepository_Integration
//
// Это платит свою цену: такие тесты проверяют КОРРЕКТНОСТЬ SQL против реальной
// схемы, поведение NULL/уникальности/транзакций — того, что mock-тесты не видят.

package main

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// testDB — общий пул для всех тестов сьюта (поднимаем контейнер один раз).
var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Поднимаем Postgres в Docker. testcontainers сам скачивает образ,
	// назначает случайный порт и ждёт готовности БД.
	container, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		panic("postgres container: " + err.Error())
	}
	// Container lifetime managed by TestMain — Terminate after the suite.
	// (testcontainers also auto-cleans via reaper if process dies.)

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("connection string: " + err.Error())
	}

	// Накатываем миграции (shared-хелпер — его использует и main.go).
	if err := runMigrations(ctx, dsn); err != nil {
		panic("migrations: " + err.Error())
	}

	// Пул для запросов приложения. В тестах берём щедрые лимиты,
	// чтобы конкурентный тест на счётчик не упёрся в размер пула.
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic("parse config: " + err.Error())
	}
	cfg.MaxConns = 20
	testDB, err = pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		panic("pgxpool: " + err.Error())
	}

	code := m.Run()

	testDB.Close()
	_ = container.Terminate(context.Background()) // best-effort cleanup
	os.Exit(code)
}

func newIntegrationRepo(t *testing.T) VideoRepository {
	t.Helper()
	require.NotNil(t, testDB, "testDB не инициализирован")
	// Чистим таблицу перед каждым тестом — изолированность.
	_, err := testDB.Exec(t.Context(), "TRUNCATE videos")
	require.NoError(t, err)
	return NewVideoRepository(testDB)
}

// Полный цикл: создаём видео, читаем, увеличиваем просмотры, проверяем.
func TestRepository_Integration_CRUD(t *testing.T) {
	repo := newIntegrationRepo(t)
	ctx := t.Context()

	created, err := repo.Create(ctx, "Введение в Go", MetadataFields{Author: "Rafael", Duration: 360})
	require.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, int64(0), created.Views)

	// jsonb вернулся — распарсим, чтобы убедиться, что данные дошли.
	var meta MetadataFields
	require.NoError(t, json.Unmarshal(created.Metadata, &meta))
	assert.Equal(t, "Rafael", meta.Author)

	got, err := repo.Get(ctx, created.ID.String())
	require.NoError(t, err)
	assert.Equal(t, created.Title, got.Title)

	// Два просмотра + проверка, что счётчик реально вырос.
	require.NoError(t, repo.AddView(ctx, created.ID.String()))
	require.NoError(t, repo.AddView(ctx, created.ID.String()))

	got, err = repo.Get(ctx, created.ID.String())
	require.NoError(t, err)
	assert.Equal(t, int64(2), got.Views)
}

// AddView для несуществующего id должен вернуть доменную ошибку.
func TestRepository_Integration_AddView_NotFound(t *testing.T) {
	repo := newIntegrationRepo(t)
	err := repo.AddView(t.Context(), "11111111-1111-1111-1111-111111111111")
	assert.ErrorIs(t, err, ErrVideoNotFound)
}

// ГОНКА СЧЁТЧИКА: 100 конкурентных AddView должны дать ровно 100.
// Атомарный UPDATE views = views + 1 справляется без блокировок.
// Если бы инкремент делался через "прочитал-добавил-записал" без FOR UPDATE,
// часть обновлений потерялась бы — это и демонстрирует ценность транзакций.
func TestRepository_Integration_AddView_Concurrent(t *testing.T) {
	repo := newIntegrationRepo(t)
	ctx := t.Context()

	v, err := repo.Create(ctx, "Concurrency demo", MetadataFields{})
	require.NoError(t, err)

	const n = 100
	var wg sync.WaitGroup
	wg.Add(n)
	errs := make(chan error, n)
	for range n {
		go func() {
			defer wg.Done()
			errs <- repo.AddView(ctx, v.ID.String())
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	got, err := repo.Get(ctx, v.ID.String())
	require.NoError(t, err)
	assert.Equal(t, int64(n), got.Views, "ожидали ровно %d просмотров", n)
}

// List возвращает видео, отсортированные по убыванию просмотров.
func TestRepository_Integration_List_Order(t *testing.T) {
	repo := newIntegrationRepo(t)
	ctx := t.Context()

	popular, err := repo.Create(ctx, "Популярное", MetadataFields{})
	require.NoError(t, err)
	rare, err := repo.Create(ctx, "Редкое", MetadataFields{})
	require.NoError(t, err)

	// Накручиваем просмотры популярному видео.
	for range 5 {
		require.NoError(t, repo.AddView(ctx, popular.ID.String()))
	}
	_ = rare // остаётся с 0 просмотров

	list, err := repo.List(ctx, 10)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(list), 2)
	assert.Equal(t, popular.ID, list[0].ID, "первым должно идти самое просматриваемое")
}
