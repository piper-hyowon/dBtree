CREATE TABLE IF NOT EXISTS sessions (
                                        id UUID PRIMARY KEY,
                                        email VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL,
    otp_code VARCHAR(10),
    otp_created_at TIMESTAMP,
    otp_expires_at TIMESTAMP,
    token VARCHAR(64),
    token_expires_at TIMESTAMP,
    resend_count INT DEFAULT 0,
    last_resend_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (email) REFERENCES users(email) ON DELETE CASCADE
    );

CREATE INDEX IF NOT EXISTS idx_sessions_email ON sessions(email);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
