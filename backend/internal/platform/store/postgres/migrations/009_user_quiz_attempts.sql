CREATE TABLE IF NOT EXISTS user_quiz_attempts (
                                                  id SERIAL PRIMARY KEY,
                                                  user_id UUID NOT NULL,
                                                  quiz_id INTEGER NOT NULL,
                                                  lemon_position_id INTEGER NOT NULL,
                                                  is_correct BOOLEAN NOT NULL,
                                                  selected_option INTEGER NOT NULL,
                                                  start_time TIMESTAMP WITH TIME ZONE NOT NULL,
                                                  submit_time TIMESTAMP WITH TIME ZONE NOT NULL,
                                                  time_taken INTEGER NOT NULL, -- 초 단위
                                                  result_status VARCHAR(20) NOT NULL,
                                                  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

                                                  CONSTRAINT fk_user
                                                      FOREIGN KEY (user_id)
                                                          REFERENCES users (id)
                                                          ON DELETE SET NULL,

                                                  CONSTRAINT fk_quiz
                                                      FOREIGN KEY (quiz_id)
                                                          REFERENCES quizzes (id)
                                                          ON DELETE SET NULL,

                                                  CONSTRAINT check_result_status
                                                      CHECK (result_status IN ('correct', 'wrong', 'timeout', 'lemon_harvested', 'harvest_timeout'))
);

CREATE INDEX IF NOT EXISTS idx_quiz_attempts_user_id ON user_quiz_attempts (user_id);
CREATE INDEX IF NOT EXISTS idx_quiz_attempts_quiz_id ON user_quiz_attempts (quiz_id);
CREATE INDEX IF NOT EXISTS idx_quiz_attempts_result ON user_quiz_attempts (result_status);
CREATE INDEX IF NOT EXISTS idx_quiz_attempts_created_at ON user_quiz_attempts (created_at);

COMMENT ON COLUMN user_quiz_attempts.result_status IS 'correct: 정답, wrong: 오답, timeout: 시간초과, lemon_harvested: 레몬 수확 성공, harvest_timeout: 레몬 수확 타임아웃';