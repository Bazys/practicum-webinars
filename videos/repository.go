package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Пакетные ошибки драйвера, чтобы различать "не найдено" и техническую ошибку.
// В pgx нет единого pgx.ErrNoRows — вместо него используется классический pgx.ErrNoRows.
var (
	// ErrVideoNotFound возвращаем, когда запись не найдена.
	// Это доменная ошибка, а не технический сбой.
	ErrVideoNotFound = errors.New("video not found")
)

// VideoRepository — контракт слоя хранения.
//
// ИНТЕРФЕЙС здесь нужен по двум причинам:
//  1. Бизнес-логику (handler) можно тестировать на моке интерфейса, без БД.
//  2. Можно подменить реализацию (например, in-memory кэш для тестов).
type VideoRepository interface {
	// List возвращает limit самых просматриваемых видео.
	List(ctx context.Context, limit int) ([]Video, error)
	// Get возвращает одно видео по id.
	Get(ctx context.Context, id string) (Video, error)
	// Create добавляет новое видео и возвращает его с присвоенным id.
	Create(ctx context.Context, title string, metadata MetadataFields) (Video, error)
	// AddView атомарно увеличивает счётчик просмотров на 1.
	// См. tx_example.go: там же показан вариант через транзакцию и SELECT FOR UPDATE.
	AddView(ctx context.Context, id string) error
}

// DBTX — минимальный набор методов для работы с БД.
// Его реализуют И настоящий пул *pgxpool.Pool, И мок pgxmock.PgxConnIface,
// поэтому репозиторий легко покрывать unit-тестами без живой БД.
//
// Обратите внимание: возвращаются типы из пакета pgx (pgx.Rows, pgx.Row),
// а не специфичные для пула — это и позволяет подменять реализацию.
type DBTX interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Убеждаемся, что *videoRepository реализует интерфейс на этапе компиляции.
var _ VideoRepository = (*videoRepository)(nil)

// videoRepository — реализация VideoRepository поверх pgxpool.
type videoRepository struct {
	db DBTX
}

// NewVideoRepository принимает *pgxpool.Pool, который реализует DBTX.
// В тестах сюда можно передать pgxmock — см. repository_pgxmock_test.go.
func NewVideoRepository(db DBTX) VideoRepository {
	return &videoRepository{db: db}
}

// List возвращает самые популярные видео.
//
// ВАЖНО: плейсхолдеры в PostgreSQL — позиционные $1, $2, …
// Привычный по MySQL/SQLite вопросительный знак "?" здесь НЕ РАБОТАЕТ
// и приведёт к ошибке синтаксиса. Это частая ловушка при переходе на Postgres.
func (r *videoRepository) List(ctx context.Context, limit int) ([]Video, error) {
	const q = `SELECT id, title, views, metadata, created_at
	           FROM videos
	           ORDER BY views DESC
	           LIMIT $1`

	rows, err := r.db.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("query videos: %w", err)
	}
	// rows обязательно закрываем — иначе соединение не вернётся в пул.
	// В pgx CollectRows закрывает rows сам, но привычка "defer Close" полезна.
	defer rows.Close()

	// pgx.CollectRows + pgx.RowToStructByName — идиоматичный способ
	// собрать *Rows в слайс структур по именам полей (совпадает по db/json-тегам
	// и именам колонок). Никакого ручного Scan.
	videos, err := pgx.CollectRows(rows, pgx.RowToStructByName[Video])
	if err != nil {
		return nil, fmt.Errorf("collect videos: %w", err)
	}
	return videos, nil
}

// Get возвращает одно видео по id.
//
// Используем QueryRow + Scan в структуру: для одной строки это проще,
// чем городить CollectRows. pgx.ErrNoRows distinguishes "нет записи".
func (r *videoRepository) Get(ctx context.Context, id string) (Video, error) {
	const q = `SELECT id, title, views, metadata, created_at
	           FROM videos
	           WHERE id = $1`

	var v Video
	err := r.db.QueryRow(ctx, q, id).Scan(
		&v.ID, &v.Title, &v.Views, &v.Metadata, &v.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Video{}, ErrVideoNotFound
	}
	if err != nil {
		return Video{}, fmt.Errorf("get video %q: %w", id, err)
	}
	return v, nil
}

// Create добавляет видео и возвращает его с присвоенным БД идентификатором.
//
// RETURNING * — postgres-специфика: получаем сгенерированные id/created_at
// без второго запроса. metadata передаём как MetadataFields — pgx сам
// упакует структуру в jsonb (см. регистрацию типа в main.go).
func (r *videoRepository) Create(ctx context.Context, title string, metadata MetadataFields) (Video, error) {
	const q = `INSERT INTO videos (title, metadata)
	           VALUES ($1, $2)
	           RETURNING id, title, views, metadata, created_at`

	var v Video
	err := r.db.QueryRow(ctx, q, title, metadata).Scan(
		&v.ID, &v.Title, &v.Views, &v.Metadata, &v.CreatedAt,
	)
	if err != nil {
		// Различаем нарушение уникальности и прочие ошибки по коду PG.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return Video{}, fmt.Errorf("create video: already exists: %w", err)
		}
		return Video{}, fmt.Errorf("create video: %w", err)
	}
	return v, nil
}

// AddView увеличивает счётчик просмотров на 1.
//
// Это АТОМАРНЫЙ вариант: инкремент выполняется одним UPDATE,
// база сама гарантирует корректность при конкурентных вызовах.
// Такой подход проще и быстрее транзакции с SELECT FOR UPDATE,
// но годится не для всех задач (см. tx_example.go).
func (r *videoRepository) AddView(ctx context.Context, id string) error {
	const q = `UPDATE videos SET views = views + 1 WHERE id = $1`

	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("add view to %q: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrVideoNotFound
	}
	return nil
}
