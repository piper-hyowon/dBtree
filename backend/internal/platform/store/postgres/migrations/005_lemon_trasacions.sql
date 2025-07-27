CREATE TABLE user_lemon_transactions
(
    id             UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    user_id        UUID                     NOT NULL,
    action_type    lemon_action             NOT NULL,
    amount         INTEGER                  NOT NULL,
    balance        INTEGER                  NOT NULL,
    status         transaction_status       NOT NULL,
    created_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    note           TEXT,
    db_instance_id VARCHAR(36)
);

-- 인덱스
CREATE INDEX idx_lemon_transactions_user_id ON user_lemon_transactions (user_id);
CREATE INDEX idx_lemon_transactions_created_at ON user_lemon_transactions (created_at);
CREATE INDEX idx_lemon_transactions_action_type ON user_lemon_transactions (action_type);
CREATE INDEX idx_lemon_transactions_status ON user_lemon_transactions (status);
CREATE INDEX idx_lemon_transactions_db_instance ON user_lemon_transactions (db_instance_id);

COMMENT ON COLUMN user_lemon_transactions.action_type IS 'harvest: 레몬 수확, instance_create: 인스턴스 생성 비용, instance_maintain: 유지 비용, welcome_bonus: 가입 보너스';
COMMENT ON COLUMN user_lemon_transactions.status IS 'successful: 성공, failed: 실패';

