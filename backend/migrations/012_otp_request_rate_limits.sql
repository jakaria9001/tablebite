CREATE TABLE IF NOT EXISTS otp_request_rate_limits (
    rate_key TEXT PRIMARY KEY,
    window_started_at TIMESTAMPTZ NOT NULL,
    request_count INTEGER NOT NULL CHECK (request_count > 0),
    last_requested_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_otp_request_rate_limits_window_started
    ON otp_request_rate_limits(window_started_at);
