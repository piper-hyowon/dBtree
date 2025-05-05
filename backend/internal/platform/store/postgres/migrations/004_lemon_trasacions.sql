CREATE TABLE IF NOT EXISTS user_lemon_transactions
(
    id             UUID,
    user_id        UUID                     NOT NULL,
    action_type    VARCHAR(20)              NOT NULL,
    amount         INTEGER                  NOT NULL,
    balance        INTEGER                  NOT NULL,
    status         VARCHAR(20)              NOT NULL,
    created_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    note           TEXT                     NULL,
    db_instance_id VARCHAR(36)              NULL,

--     CONSTRAINT fk_transactions_user
--         FOREIGN KEY (user_id)
--             REFERENCES users (id)
--             ON DELETE SET NULL,

    CONSTRAINT check_action_type
        CHECK (action_type IN ('harvest', 'instance_create', 'instance_maintain', 'welcome_bonus')),

    CONSTRAINT check_status
        CHECK (status IN ('successful', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_lemon_transactions_user_id ON user_lemon_transactions (user_id);
CREATE INDEX IF NOT EXISTS idx_lemon_transactions_created_at ON user_lemon_transactions (created_at);
CREATE INDEX IF NOT EXISTS idx_lemon_transactions_action_type ON user_lemon_transactions (action_type);
CREATE INDEX IF NOT EXISTS idx_lemon_transactions_status ON user_lemon_transactions (status);
CREATE INDEX IF NOT EXISTS idx_lemon_transactions_db_instance ON user_lemon_transactions (db_instance_id);


COMMENT ON COLUMN user_lemon_transactions.action_type IS 'harvest: 레몬 수확, instance_create: 인스턴스 생성 비용, instance_maintain: 유지 비용, welcome_bonus: 가입 보너스';
COMMENT ON COLUMN user_lemon_transactions.status IS 'successful: 성공, failed: 실패';

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS lemon_balance   INTEGER                  NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_harvest_at TIMESTAMP WITH TIME ZONE NULL;
