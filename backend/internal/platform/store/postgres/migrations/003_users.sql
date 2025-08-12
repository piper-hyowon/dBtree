DROP TABLE IF EXISTS users CASCADE;

CREATE TABLE users
(
    id                  UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    email               VARCHAR(255) UNIQUE      NOT NULL,
    is_deleted          BOOLEAN                  NOT NULL DEFAULT FALSE,
    welcome_bonus_given BOOLEAN                  NOT NULL DEFAULT FALSE,

    -- 레몬 관련
    lemon_balance       INTEGER                  NOT NULL DEFAULT 0,
    total_earned_lemons BIGINT                   NOT NULL DEFAULT 0,
    total_spent_lemons  BIGINT                   NOT NULL DEFAULT 0,
    last_harvest_at     TIMESTAMP WITH TIME ZONE,

    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique ON users (email) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_users_is_deleted ON users (is_deleted);
CREATE INDEX IF NOT EXISTS idx_users_welcome_bonus ON users (welcome_bonus_given) WHERE welcome_bonus_given = FALSE;

CREATE TRIGGER update_users_timestamp
    BEFORE UPDATE
    ON users
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp_column();