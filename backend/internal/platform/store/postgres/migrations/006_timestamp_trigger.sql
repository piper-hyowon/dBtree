-- 타임스탬프 자동 업데이트를 위한 트리거 함수
CREATE OR REPLACE FUNCTION update_timestamp_column()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 트리거 적용
CREATE TRIGGER update_lemons_timestamp
    BEFORE UPDATE
    ON lemons
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp_column();

CREATE TRIGGER update_users_timestamp
    BEFORE UPDATE
    ON users
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp_column();

CREATE TRIGGER update_users_timestamp
    BEFORE UPDATE
    ON sessions
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp_column();

