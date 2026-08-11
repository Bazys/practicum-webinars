package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DSN по умолчанию совпадает с docker-compose.yaml (user/pass/db = postgres).
// В проде DSN обязательно берётся из конфига/env, никогда не хардкодится.
const defaultDSN = "postgres://postgres:postgres@localhost:5432/my_database?sslmode=disable"

func main() {
	dsn := envOr("DATABASE_URL", defaultDSN)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 1. Пул соединений.
	//
	// pgxpool.NewWithConfig позволяет настроить лимиты — это КРИТИЧНО для прода:
	//   - MaxConns         — верхняя граница одновременных соединений (по умолчанию max(4, runtime.NumCPU));
	//   - MinConns         — "тёплые" соединения, которые держим всегда;
	//   - MaxConnLifetime  — как часто пересоздавать соединение (защита от протухания);
	//   - MaxConnIdleTime  — когда закрывать простаивающее соединение.
	// Если лимиты не задать, под нагрузкой можно либо исчерпать соединения в БД,
	// либо наоборот держать слишком много "лишних".
	pool, err := newPool(ctx, dsn)
	if err != nil {
		log.Fatalf("pgxpool: %v", err)
	}
	defer pool.Close()

	// 2. Миграции (накатываются при старте приложения).
	if err := runMigrations(ctx, dsn); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	// 3. Приложение.
	repo := NewVideoRepository(pool)
	api := NewAPI(repo)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           api.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Print("listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	// 4. Graceful shutdown: ждём сигнала и аккуратно гасим сервер и пул.
	<-ctx.Done()
	log.Print("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown: %v", err)
	}
}

// newPool создаёт настроенный пул соединений и проверяет подключение.
func newPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Пингуем с таймаутом, чтобы сразу падать, если БД недоступна.
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
