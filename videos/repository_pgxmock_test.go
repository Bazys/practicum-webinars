package main

// Этот файл — UNIT-тесты репозитория на pgxmock (https://github.com/pashagolub/pgxmock).
//
// В отличие от тестов handler_test.go (где мы мокали доменный интерфейс),
// здесь мы мокаем САМ ДРАЙВЕР pgx. Это позволяет проверить:
//   - точный текст SQL;
//   - правильные плейсхолдеры ($1, а не ?);
//   - порядок и значения аргументов;
//   - корректную обработку ошибок БД (включая pgx.ErrNoRows).
//
// Платой за детализацию является хрупкость: тест жёстко привязан к тексту запроса.
// Это нормально — если кто-то поменяет запрос, тест должен сообщить об этом.
// Но реальную корректность SQL против схемы проверяют ИНТЕГРАЦИОННЫЕ тесты
// (см. repository_integration_test.go).

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockRepo(t *testing.T) (VideoRepository, pgxmock.PgxConnIface) {
	t.Helper()
	mock, err := pgxmock.NewConn(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherRegexp))
	require.NoError(t, err)
	// PgxConnIface реализует наш DBTX (Query/QueryRow/Exec), поэтому мок
	// можно передать прямо в конструктор репозитория — никакой БД не нужно.
	return NewVideoRepository(mock), mock
}

func TestRepository_List(t *testing.T) {
	repo, mock := newMockRepo(t)
	defer mock.Close(t.Context())

	// Ожидаем SQL с плейсхолдером $1 и лимитом в качестве аргумента.
	rows := mock.NewRows([]string{"id", "title", "views", "metadata", "created_at"}).
		AddRow(pgtype.UUID{Bytes: [16]byte{1}, Valid: true}, "Go 101", int64(5),
			[]byte(`{}`), time.Now())

	mock.ExpectQuery(`SELECT id, title, views, metadata, created_at\s+FROM videos\s+ORDER BY views DESC\s+LIMIT \$1`).
		WithArgs(10).
		WillReturnRows(rows)

	videos, err := repo.List(t.Context(), 10)
	require.NoError(t, err)
	require.Len(t, videos, 1)
	assert.Equal(t, "Go 101", videos[0].Title)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Get_NotFound(t *testing.T) {
	repo, mock := newMockRepo(t)
	defer mock.Close(t.Context())

	// Имитируем отсутствие строки — pgx вернёт pgx.ErrNoRows.
	mock.ExpectQuery(`SELECT .* FROM videos WHERE id = \$1`).
		WithArgs("missing").
		WillReturnError(pgx.ErrNoRows)

	_, err := repo.Get(t.Context(), "missing")
	assert.ErrorIs(t, err, ErrVideoNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_AddView_AffectedRows(t *testing.T) {
	repo, mock := newMockRepo(t)
	defer mock.Close(t.Context())

	// UPDATE должен затронуть ровно одну строку.
	mock.ExpectExec(`UPDATE videos SET views = views \+ 1 WHERE id = \$1`).
		WithArgs("11111111-1111-1111-1111-111111111111").
		WillReturnResult(pgconn.NewCommandTag("UPDATE 1"))

	err := repo.AddView(t.Context(), "11111111-1111-1111-1111-111111111111")
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_AddView_NotFound(t *testing.T) {
	repo, mock := newMockRepo(t)
	defer mock.Close(t.Context())

	// Ни одна строка не обновлена → видео не существует.
	mock.ExpectExec(`UPDATE videos SET views = views \+ 1 WHERE id = \$1`).
		WithArgs("missing").
		WillReturnResult(pgconn.NewCommandTag("UPDATE 0"))

	err := repo.AddView(t.Context(), "missing")
	assert.ErrorIs(t, err, ErrVideoNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Create_UniqueViolation(t *testing.T) {
	repo, mock := newMockRepo(t)
	defer mock.Close(t.Context())

	// Имитируем нарушение уникальности — код PG 23505.
	mock.ExpectQuery(`INSERT INTO videos`).
		WithArgs("dup", MetadataFields{}).
		WillReturnError(&pgconn.PgError{Code: "23505"})

	_, err := repo.Create(t.Context(), "dup", MetadataFields{})
	require.Error(t, err)
	// Реализация оборачивает ошибку, проверим что внутри — нарушение уникальности.
	var pgErr *pgconn.PgError
	require.True(t, errors.As(err, &pgErr))
	assert.Equal(t, "23505", pgErr.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
