ALTER TABLE todos ADD COLUMN status INT NOT NULL DEFAULT 1;
UPDATE todos SET status = 4 WHERE completed = true;
ALTER TABLE todos DROP COLUMN completed;
DROP INDEX IF EXISTS idx_todos_completed_id;
CREATE INDEX IF NOT EXISTS idx_todos_status_id ON todos (status ASC, id DESC);
