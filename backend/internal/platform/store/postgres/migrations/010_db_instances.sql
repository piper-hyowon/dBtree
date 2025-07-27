CREATE TABLE IF NOT EXISTS db_instances
(
    id                    BIGSERIAL PRIMARY KEY,
    external_id           UUID                                         DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    user_id               UUID                                NOT NULL REFERENCES users (id),

    -- 기본 정보
    name                  VARCHAR(255)                        NOT NULL,
    type                  db_type                             NOT NULL,
    size                  db_size                             NOT NULL,
    mode                  VARCHAR(50)                         NOT NULL,

    -- 프리셋 참조 (통계용)
    created_from_preset   VARCHAR(100),

    -- 리소스 (프리셋에서 복사되거나 커스텀 입력)
    cpu                   INTEGER CHECK (cpu > 0)             NOT NULL,
    memory                INTEGER CHECK (memory > 0)          NOT NULL,            -- MB
    disk                  INTEGER CHECK (disk > 0)            NOT NULL,            -- GB

    -- 비용
    creation_cost         INTEGER CHECK (creation_cost >= 0)  NOT NULL,
    hourly_cost           INTEGER CHECK (hourly_cost >= 0)    NOT NULL,
    minimum_lemons        INTEGER CHECK (minimum_lemons >= 0) NOT NULL DEFAULT 24, -- 최소 24시간

    -- 상태
    status                db_status                           NOT NULL DEFAULT 'provisioning',
    status_reason         TEXT,

    -- K8s 정보
    k8s_namespace         VARCHAR(255),
    k8s_resource_name     VARCHAR(255),

    -- 연결 정보
    endpoint              VARCHAR(255),
    port                  INTEGER CHECK (port > 0 AND port <= 65535),
    password_encrypted    TEXT,

    -- 설정 (프리셋 기본값 + 사용자 커스텀)
    config                JSONB                               NOT NULL DEFAULT '{}',

    -- 백업 설정
    backup_enabled        BOOLEAN                             NOT NULL DEFAULT false,
    backup_schedule       VARCHAR(100),                                            -- cron format
    backup_retention_days INTEGER CHECK (backup_retention_days > 0),

    -- 시간
    created_at            TIMESTAMP WITH TIME ZONE            NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMP WITH TIME ZONE            NOT NULL DEFAULT NOW(),
    last_billed_at        TIMESTAMP WITH TIME ZONE,
    paused_at             TIMESTAMP WITH TIME ZONE,
    deleted_at            TIMESTAMP WITH TIME ZONE,

    CONSTRAINT unique_user_instance_name UNIQUE (user_id, name, deleted_at)
);

CREATE INDEX idx_db_instances_user_id ON db_instances (user_id);
CREATE INDEX idx_db_instances_status ON db_instances (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_db_instances_billing ON db_instances (last_billed_at)
    WHERE status = 'running' AND deleted_at IS NULL;

CREATE TRIGGER update_db_instances_timestamp
    BEFORE UPDATE
    ON db_instances
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp_column();