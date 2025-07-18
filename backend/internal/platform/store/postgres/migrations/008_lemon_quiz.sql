CREATE TABLE IF NOT EXISTS lemon_quiz (
                                                   id SERIAL PRIMARY KEY,
                                                   lemon_position_id INTEGER NOT NULL,
                                                   quiz_id INTEGER NOT NULL,
                                                   created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                                                   updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

                                                   CONSTRAINT fk_lemon_position
                                                       FOREIGN KEY (lemon_position_id)
                                                           REFERENCES lemons (position_id)
                                                           ON DELETE CASCADE,

                                                   CONSTRAINT fk_quiz
                                                       FOREIGN KEY (quiz_id)
                                                           REFERENCES quizzes (id)
                                                           ON DELETE CASCADE,

                                                   CONSTRAINT unique_lemon_position
                                                       UNIQUE (lemon_position_id)
);

CREATE INDEX IF NOT EXISTS idx_lemon_quiz_quiz_id ON lemon_quiz (quiz_id);

COMMENT ON TABLE lemon_quiz IS '레몬과 퀴즈의 1:1 매핑 테이블';



CREATE TRIGGER update_lemon_quiz_timestamp
    BEFORE UPDATE
    ON lemon_quiz
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp_column();

