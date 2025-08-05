CREATE TABLE IF NOT EXISTS lemon_refund_failures
(
    id            SERIAL PRIMARY KEY,
    user_id       UUID    NOT NULL,
    amount        INTEGER NOT NULL,
    reason        TEXT,
    error_message TEXT,
    resolved      BOOLEAN                  DEFAULT FALSE,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    resolved_at   TIMESTAMP WITH TIME ZONE
);