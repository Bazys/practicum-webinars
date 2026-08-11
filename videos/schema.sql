-- Демо-данные для ручного запуска.
-- Применяется СТРОГО ПОСЛЕ миграций, например так:
--   psql ... -f migrations/0001_init.up.sql && psql ... -f schema.sql
-- В автотестах этот файл НЕ используется — там данные готовит сам тест.
INSERT INTO videos (title, views, metadata)
VALUES
    ('Введение в Go',        42,  '{"author": "Rafael", "duration": 360}'::jsonb),
    ('Контекст в Go',         15,  '{"author": "Rafael", "duration": 720}'::jsonb),
    ('Работа с PostgreSQL',   7,   '{"author": "Rafael", "duration": 1800}'::jsonb)
ON CONFLICT DO NOTHING;
