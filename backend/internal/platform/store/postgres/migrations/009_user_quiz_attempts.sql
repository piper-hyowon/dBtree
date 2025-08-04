CREATE TABLE IF NOT EXISTS user_quiz_attempts
(
    id                 SERIAL PRIMARY KEY,
    user_id            UUID                     NOT NULL REFERENCES users (id) ON DELETE SET NULL,
    quiz_id            INTEGER                  NOT NULL REFERENCES quizzes (id) ON DELETE SET NULL,
    lemon_position_id  INTEGER                  NOT NULL,

    -- 결과
    is_correct         BOOLEAN,
    selected_option    INTEGER,

    -- 시간
    start_time         TIMESTAMP WITH TIME ZONE NOT NULL,
    submit_time        TIMESTAMP WITH TIME ZONE,
    time_taken         INTEGER, -- 초
    time_taken_clicked INTEGER,

    -- 상태
    status             quiz_status              NOT NULL DEFAULT 'started',
    harvest_status     harvest_status           NOT NULL DEFAULT 'none',

    created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);


CREATE INDEX IF NOT EXISTS idx_quiz_attempts_user_id ON user_quiz_attempts (user_id);
CREATE INDEX IF NOT EXISTS idx_quiz_attempts_quiz_id ON user_quiz_attempts (quiz_id);
CREATE INDEX IF NOT EXISTS idx_quiz_attempts_result ON user_quiz_attempts (status);
CREATE INDEX IF NOT EXISTS idx_quiz_attempts_created_at ON user_quiz_attempts (created_at);

DROP TRIGGER IF EXISTS update_user_quiz_attempts_timestamp ON user_quiz_attempts;
CREATE TRIGGER update_user_quiz_attempts_timestamp
    BEFORE UPDATE
    ON user_quiz_attempts
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp_column();
