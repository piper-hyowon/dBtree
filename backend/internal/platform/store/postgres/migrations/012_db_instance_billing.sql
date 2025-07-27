CREATE TABLE IF NOT EXISTS db_instance_billing
(
    id                  BIGSERIAL PRIMARY KEY,
    db_instance_id      BIGINT                           NOT NULL REFERENCES db_instances (id) ON DELETE CASCADE,

    -- 빌링 주기
    billing_cycle_start TIMESTAMP WITH TIME ZONE         NOT NULL,
    billing_cycle_end   TIMESTAMP WITH TIME ZONE         NOT NULL,

    -- 비용
    hourly_rate         INTEGER CHECK (hourly_rate > 0)  NOT NULL,           -- 시간당 비용
    hours               INTEGER                          NOT NULL DEFAULT 1, -- 청구 시간
    total_amount        INTEGER CHECK (total_amount > 0) NOT NULL,           -- 총 비용

    -- 처리 상태
    status              billing_status                   NOT NULL DEFAULT 'pending',

    -- 처리 결과
    processed_at        TIMESTAMP WITH TIME ZONE,
    transaction_id      UUID REFERENCES user_lemon_transactions (id),
    failure_reason      TEXT,
    retry_count         INTEGER                                   DEFAULT 0,

    created_at          TIMESTAMP WITH TIME ZONE         NOT NULL DEFAULT NOW(),

    -- 중복 방지
    CONSTRAINT unique_billing_cycle
        UNIQUE (db_instance_id, billing_cycle_start)
);

-- 인덱스
CREATE INDEX idx_billing_pending
    ON db_instance_billing (billing_cycle_start)
    WHERE status = 'pending';

CREATE INDEX idx_billing_instance
    ON db_instance_billing (db_instance_id);