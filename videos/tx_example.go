package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Этот файл — УЧЕБНЫЙ пример транзакций в pgx.
//
// В repository.go AddView сделан одним атомарным UPDATE — это правильный
// способ для счётчика. Но не всякая задача сводится к одному запросу.
// Здесь показан общий шаблон "прочитал-вычислил-записал", где без блокировки
// строки при конкурентной нагрузке легко получить потерю данных.

// AddViewInTransaction — то же, что AddView, но через явную транзакцию.
//
// Шаги:
//  1. Begin — открываем транзакцию (по умолчанию READ COMMITTED).
//  2. SELECT ... FOR UPDATE — читаем текущее значение views И блокируем строку.
//     Пока транзакция не завершится, никто другой не сможет изменить эту строку
//     (его UPDATE будет ждать).
//  3. UPDATE — записываем новое значение.
//  4. Commit — фиксируем изменения и снимаем блокировку.
//
// defer Rollback — страховка: если что-то пошло не так до Commit,
// транзакция откатится. Если Commit уже прошёл, Rollback просто вернёт ошибку,
// которую мы игнорируем.
//
// ПОЧЕМУ для счётчика это ХУЖЕ атомарного UPDATE:
//   - два круглых пути в БД вместо одного;
//   - строка заблокирована дольше → меньше параллелизма под нагрузкой.
//
// Зато шаблон универсален: если бы новое значение зависело от сложной
// бизнес-логики (а не от +1), атомарный UPDATE бы не спас.
func AddViewInTransaction(ctx context.Context, db *pgxpool.Pool, id string) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	// Откат при ошибке. Commit ниже завершает tx, и тогда Rollback вернёт
	// sql.ErrTxDone — это нормально, ошибку игнорируем.
	defer func() { _ = tx.Rollback(ctx) }()

	var views int64
	// FOR UPDATE — ключевое слово. Без него две конкурентные транзакции
	// прочитали бы одинаковое значение views и потеряли один просмотр.
	err = tx.QueryRow(ctx,
		`SELECT views FROM videos WHERE id = $1 FOR UPDATE`, id,
	).Scan(&views)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrVideoNotFound
	}
	if err != nil {
		return fmt.Errorf("select views for update: %w", err)
	}

	_, err = tx.Exec(ctx,
		`UPDATE videos SET views = $1 WHERE id = $2`, views+1, id,
	)
	if err != nil {
		return fmt.Errorf("update views: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
