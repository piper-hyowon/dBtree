CREATE TABLE IF NOT EXISTS db_presets
(
    id                   VARCHAR(100) PRIMARY KEY,
    type                 db_type                            NOT NULL,
    size                 db_size                            NOT NULL,
    mode                 db_mode                            NOT NULL,

    -- 표시 정보
    name                 VARCHAR(255)                       NOT NULL,
    icon                 VARCHAR(50),
    description          TEXT                               NOT NULL,
    friendly_description TEXT,
    technical_terms      JSONB,
    use_cases            TEXT[]                             NOT NULL,

    -- 리소스 & 비용
    cpu                  DECIMAL(4, 2) CHECK (cpu > 0)      NOT NULL, -- 변경: INTEGER → DECIMAL
    memory               INTEGER CHECK (memory > 0)         NOT NULL,
    disk                 INTEGER CHECK (disk > 0)           NOT NULL,
    creation_cost        INTEGER CHECK (creation_cost >= 0) NOT NULL,
    hourly_cost          INTEGER CHECK (hourly_cost >= 0)   NOT NULL,

    -- 기본 설정
    default_config       JSONB                              NOT NULL,

    -- 정렬 & 활성화
    sort_order           INTEGER                            NOT NULL DEFAULT 0,
    is_active            BOOLEAN                            NOT NULL DEFAULT true,

    created_at           TIMESTAMP WITH TIME ZONE           NOT NULL DEFAULT NOW()
);