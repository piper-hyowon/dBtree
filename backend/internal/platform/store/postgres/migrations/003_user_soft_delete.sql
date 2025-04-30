ALTER TABLE users ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT FALSE;

DROP INDEX IF EXISTS users_email_unique_idx;
CREATE UNIQUE INDEX users_email_unique_idx ON users (email) WHERE is_deleted = FALSE;

CREATE INDEX users_is_deleted_idx ON users (is_deleted);