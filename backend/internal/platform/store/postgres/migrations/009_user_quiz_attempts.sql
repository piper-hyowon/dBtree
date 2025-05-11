CREATE TABLE IF NOT EXISTS user_quiz_attempts
(
    id                 SERIAL PRIMARY KEY,
    user_id            UUID                     NOT NULL,
    quiz_id            INTEGER                  NOT NULL,
    lemon_position_id  INTEGER                  NOT NULL,
    is_correct         BOOLEAN                  NULL,                       -- 정답 여부
    selected_option    INTEGER                  NULL,
    start_time         TIMESTAMP WITH TIME ZONE NOT NULL,
    submit_time        TIMESTAMP WITH TIME ZONE NULL,
    time_taken         INTEGER                  NULL,                       -- 초 단위
    time_taken_clicked INTEGER                  NULL,

    status             VARCHAR(20)              NOT NULL DEFAULT 'started', -- 퀴즈 진행 상태
    harvest_status     varchar(20)              NOT NULL DEFAULT 'none',    -- 수확 단계 상태
    created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_user
        FOREIGN KEY (user_id)
            REFERENCES users (id)
            ON DELETE SET NULL,

    CONSTRAINT fk_quiz
        FOREIGN KEY (quiz_id)
            REFERENCES quizzes (id)
            ON DELETE SET NULL,

    CONSTRAINT check_status
        CHECK (status IN
               ('started', 'done', 'timeout')),
    CONSTRAINT check_harvest_status
        CHECK (harvest_status IN
               ('none', 'in_progress', 'success', 'timeout'))
);

CREATE INDEX IF NOT EXISTS idx_quiz_attempts_user_id ON user_quiz_attempts (user_id);
CREATE INDEX IF NOT EXISTS idx_quiz_attempts_quiz_id ON user_quiz_attempts (quiz_id);
CREATE INDEX IF NOT EXISTS idx_quiz_attempts_result ON user_quiz_attempts (status);
CREATE INDEX IF NOT EXISTS idx_quiz_attempts_created_at ON user_quiz_attempts (created_at);

CREATE TRIGGER update_user_quiz_attempts_timestamp
    BEFORE UPDATE
    ON user_quiz_attempts
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp_column();
