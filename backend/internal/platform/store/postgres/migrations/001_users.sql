CREATE TABLE IF NOT EXISTS users
(
    id
               UUID
        PRIMARY
            KEY,
    email
               VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP           NOT NULL DEFAULT NOW
                                                    (
                                                    ),
    updated_at TIMESTAMP           NOT NULL DEFAULT NOW
                                                    (
                                                    )
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);