package main

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Video — доменная модель видео.
//
// Здесь намеренно используются типы, специфичные для PostgreSQL:
//
//   - pgtype.UUID ↔ колонка uuid. Корректно отличает NULL от нулевого значения.
//   - json.RawMessage (Metadata) ↔ колонка jsonb. pgx умеет кодировать/декодировать
//     jsonb в []byte/json.RawMessage напрямую, без специальных типов-обёрток.
//     При желании jsonb можно мапить и в структуру — см. MetadataFields и
//     регистрацию кодека в тестах/приложении.
//
// Для CreatedAt берём обычный time.Time: pgx сканирует timestamptz в него напрямую.
type Video struct {
	ID        pgtype.UUID     `json:"id"`
	Title     string          `json:"title"`
	Views     int64           `json:"views"`
	Metadata  json.RawMessage `json:"metadata"`
	CreatedAt time.Time       `json:"created_at"`
}

// MetadataFields — распакованный JSON из колонки metadata.
// Показывает, что jsonb можно читать не только сырыми байтами, но и в структуру.
type MetadataFields struct {
	Author   string `json:"author"`
	Duration int    `json:"duration"`
}
