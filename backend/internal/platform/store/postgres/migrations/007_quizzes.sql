CREATE TABLE IF NOT EXISTS quizzes (
                                       id SERIAL PRIMARY KEY,
                                       question TEXT NOT NULL,
                                       options JSONB NOT NULL,
                                       correct_option_idx INTEGER NOT NULL,
                                       difficulty VARCHAR(20) NOT NULL,
                                       category VARCHAR(50) NOT NULL,
                                       explanation TEXT,
                                       time_limit INTEGER NOT NULL,
                                       is_active BOOLEAN NOT NULL DEFAULT FALSE,
                                       usage_count INTEGER NOT NULL DEFAULT 0,
                                       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                                       updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

                                       CONSTRAINT check_correct_option_valid CHECK (
                                           (jsonb_array_length(options) > correct_option_idx) AND (correct_option_idx >= 0)
                                           ),

                                       CONSTRAINT check_difficulty
                                           CHECK (difficulty IN ('easy', 'normal')),

                                       CONSTRAINT check_category
                                           CHECK (category IN ('basics', 'sql', 'design'))
);

CREATE INDEX IF NOT EXISTS idx_quizzes_difficulty ON quizzes (difficulty);
CREATE INDEX IF NOT EXISTS idx_quizzes_category ON quizzes (category);
CREATE INDEX IF NOT EXISTS idx_quizzes_is_active ON quizzes (is_active);
CREATE INDEX IF NOT EXISTS idx_quizzes_usage_count ON quizzes (usage_count);

COMMENT ON COLUMN quizzes.correct_option_idx IS '정답 선택지 인덱스 (0부터 시작)';
COMMENT ON COLUMN quizzes.time_limit IS '퀴즈 제한 시간(초)';
COMMENT ON COLUMN quizzes.is_active IS '현재 레몬에 할당되어 있는지 여부';
COMMENT ON COLUMN quizzes.usage_count IS '퀴즈가 사용된 횟수, 랜덤 선택 시 활용';