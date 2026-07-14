-- This script initializes the database table for the Todo app.
-- You can run this in your psql terminal or PgAdmin.

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS todos (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    status INT NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS trgm_idx_todos_title ON todos USING gin (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_todos_status_id ON todos (status ASC, id DESC);
