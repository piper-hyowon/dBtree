CREATE TABLE IF NOT EXISTS db_instance_backups
(
    id            SERIAL PRIMARY KEY,
    instance_id   BIGINT                   NOT NULL REFERENCES db_instances (id) ON DELETE CASCADE,
    external_id   VARCHAR(36)              NOT NULL UNIQUE,
    name          VARCHAR(255)             NOT NULL,
    type          VARCHAR(50)              NOT NULL CHECK (type IN ('manual', 'scheduled')),
    status        VARCHAR(50)              NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    k8s_job_name  VARCHAR(255),
    size_bytes    BIGINT,
    storage_path  TEXT,
    error_message TEXT,
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at  TIMESTAMP WITH TIME ZONE,
    expires_at    TIMESTAMP WITH TIME ZONE
);

CREATE TRIGGER update_db_instance_backups_timestamp
    BEFORE UPDATE
    ON db_instance_backups
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp_column();

CREATE INDEX IF NOT EXISTS idx_backups_instance_id
    ON db_instance_backups (instance_id);

CREATE INDEX IF NOT EXISTS idx_backups_status
    ON db_instance_backups (status);

CREATE INDEX IF NOT EXISTS idx_backups_created_at
    ON db_instance_backups (created_at);