CREATE TABLE IF NOT EXISTS waiter_staff (
    id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT NOT NULL UNIQUE REFERENCES customers(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS waiter_sessions (
    id BIGSERIAL PRIMARY KEY,
    waiter_id BIGINT NOT NULL REFERENCES waiter_staff(id) ON DELETE CASCADE,
    session_token_hash BYTEA NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    CONSTRAINT waiter_sessions_token_hash_length CHECK (octet_length(session_token_hash) = 32)
);

CREATE INDEX IF NOT EXISTS idx_waiter_sessions_active_expiry
    ON waiter_sessions(expires_at)
    WHERE revoked_at IS NULL;

ALTER TABLE order_status_history
    ADD COLUMN IF NOT EXISTS waiter_id BIGINT REFERENCES waiter_staff(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_order_status_history_waiter
    ON order_status_history(waiter_id, changed_at DESC);
