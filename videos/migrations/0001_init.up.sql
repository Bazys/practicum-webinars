-- Создание таблицы видео.
-- Обратите внимание на типы, специфичные для PostgreSQL:
--   uuid   — компактный 128-битный идентификатор
--   jsonb  — бинарный JSON, по которому можно строить индексы и делать запросы
--   timestamptz — момент времени с учётом часового пояса
CREATE TABLE IF NOT EXISTS videos (
    id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    title      text        NOT NULL,
    views      bigint      NOT NULL DEFAULT 0,
    -- metadata — произвольный JSON: теги, длительность, автор и т.п.
    metadata   jsonb       NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_videos_views ON videos (views);
