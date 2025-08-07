-- 레몬 거래시 사용자의 total_earned_lemons, total_spent_lemons 자동 업데이트

-- 트리거 함수 생성
CREATE OR REPLACE FUNCTION update_user_lemon_totals()
    RETURNS TRIGGER AS
$$
BEGIN
    -- successful 트랜잭션만 처리
    IF NEW.status = 'successful' THEN
        IF NEW.amount > 0 THEN
            -- 양수 amount는 earned에 추가
            UPDATE users
            SET total_earned_lemons = total_earned_lemons + NEW.amount
            WHERE id = NEW.user_id;
        ELSE
            -- 음수 amount는 spent에 추가 (절댓값으로)
            UPDATE users
            SET total_spent_lemons = total_spent_lemons + ABS(NEW.amount)
            WHERE id = NEW.user_id;
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 기존 트리거가 있다면 삭제
DROP TRIGGER IF EXISTS trigger_update_user_lemon_totals ON user_lemon_transactions;

-- 트리거 생성
CREATE TRIGGER trigger_update_user_lemon_totals
    AFTER INSERT
    ON user_lemon_transactions
    FOR EACH ROW
EXECUTE FUNCTION update_user_lemon_totals();

-- 기존 데이터로 total 값 초기화 (이미 실행된 경우 주석 처리)
-- UPDATE users
-- SET total_earned_lemons = (SELECT COALESCE(SUM(amount), 0)
--                            FROM user_lemon_transactions
--                            WHERE user_id = users.id
--                              AND status = 'successful'
--                              AND amount > 0),
--     total_spent_lemons  = (SELECT COALESCE(SUM(ABS(amount)), 0)
--                            FROM user_lemon_transactions
--                            WHERE user_id = users.id
--                              AND status = 'successful'
--                              AND amount < 0)
-- WHERE id IS NOT NULL;

-- 필요시 수동으로 실행할 수 있도록 보관l