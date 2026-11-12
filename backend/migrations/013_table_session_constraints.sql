WITH active_sessions AS (
    SELECT id,
           ROW_NUMBER() OVER (PARTITION BY table_id ORDER BY opened_at DESC, id DESC) AS position
    FROM table_sessions
    WHERE closed_at IS NULL
)
UPDATE table_sessions
SET closed_at = NOW()
WHERE id IN (SELECT id FROM active_sessions WHERE position > 1);

CREATE UNIQUE INDEX IF NOT EXISTS idx_table_sessions_one_active_per_table
    ON table_sessions(table_id)
    WHERE closed_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_table_sessions_table_closed
    ON table_sessions(table_id, closed_at, opened_at DESC);
