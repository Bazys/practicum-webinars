package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	// Стандартный драйвер pgx для database/sql — нужен migrate-драйверу.
	// Само приложение ходит в БД через нативный pgxpool, а миграции — через sql-интерфейс.
	_ "github.com/jackc/pgx/v5/stdlib"
)

// migrations встраивает каталог migrations/ прямо в бинарник.
// Тогда для наката схемы не нужны внешние файлы: переехал бинарник — миграции с ним.
// embed.FS реализует fs.FS / fs.ReadDirFS, что и требует iofs-источник migrate.
//
//go:embed migrations
var migrations embed.FS

// runMigrations накатывает все доступные миграции.
// dsn — строка вида "postgres://user:pass@host:5432/db?sslmode=disable".
func runMigrations(_ context.Context, dsn string) error {
	src, err := iofs.New(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("migrate source: %w", err)
	}

	// migrate гоняется через database/sql, а не через нативный pgxpool:
	// драйвер pgx5 (pgx/v5/stdlib) оборачивает pgx в интерфейс sql.DB.
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open for migrate: %w", err)
	}
	defer db.Close()

	// pgxmigrate.WithInstance превращает *sql.DB в database.Driver для migrate.
	dbDriver, err := pgxmigrate.WithInstance(db, &pgxmigrate.Config{})
	if err != nil {
		return fmt.Errorf("migrate driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "pgx5", dbDriver)
	if err != nil {
		return fmt.Errorf("migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
