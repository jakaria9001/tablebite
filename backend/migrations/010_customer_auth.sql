CREATE TABLE IF NOT EXISTS customers (
    id BIGSERIAL PRIMARY KEY,
    phone_number TEXT NOT NULL,
    phone_verified BOOLEAN NOT NULL DEFAULT FALSE,
    name TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ,
    CONSTRAINT customers_phone_number_unique UNIQUE (phone_number),
    CONSTRAINT customers_phone_number_format CHECK (phone_number ~ '^\+91[6-9][0-9]{9}$')
);

CREATE TABLE IF NOT EXISTS otp_challenges (
    id BIGSERIAL PRIMARY KEY,
    phone_number TEXT NOT NULL,
    otp_hash BYTEA NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 5,
    verified_at TIMESTAMPTZ,
    consumed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT otp_challenges_phone_number_format CHECK (phone_number ~ '^\+91[6-9][0-9]{9}$'),
    CONSTRAINT otp_challenges_otp_hash_length CHECK (octet_length(otp_hash) = 32),
    CONSTRAINT otp_challenges_attempt_count_valid CHECK (attempt_count >= 0 AND attempt_count <= max_attempts),
    CONSTRAINT otp_challenges_max_attempts_valid CHECK (max_attempts > 0)
);

CREATE INDEX IF NOT EXISTS idx_otp_challenges_phone_created
    ON otp_challenges(phone_number, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_otp_challenges_active_phone
    ON otp_challenges(phone_number, created_at DESC)
    WHERE consumed_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_otp_challenges_expires_at
    ON otp_challenges(expires_at);

CREATE TABLE IF NOT EXISTS customer_sessions (
    id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    session_token_hash BYTEA NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    CONSTRAINT customer_sessions_token_hash_unique UNIQUE (session_token_hash),
    CONSTRAINT customer_sessions_token_hash_length CHECK (octet_length(session_token_hash) = 32)
);

CREATE INDEX IF NOT EXISTS idx_customer_sessions_customer
    ON customer_sessions(customer_id, expires_at DESC);

CREATE INDEX IF NOT EXISTS idx_customer_sessions_active_expiry
    ON customer_sessions(expires_at)
    WHERE revoked_at IS NULL;
